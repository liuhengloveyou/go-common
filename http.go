package common

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// 上传接口把文件放在哪个目录
var UploadDir string

// 文件下载
const ONEPIECE = 10 * 1024
const (
	MD5MODEALL = iota
	MD5MODE1M  /*文件长度+每隔1M取开始10K+文件尾10K*/
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64*1024)
	},
}

func Md51m(f *os.File) (string, error) {
	m := md5.New()
	b := make([]byte, ONEPIECE)

	fs, err := f.Stat()
	if err != nil {
		return "", err
	}

	io.WriteString(m, fmt.Sprintf("%d", fs.Size()))

	var i int64
	for ; i <= (fs.Size() / (100 * ONEPIECE)); i++ {
		f.Seek(i*100*ONEPIECE, 0)
		n, e := f.Read(b)
		if e != nil && e != io.EOF {
			return "", e
		}
		if n > 0 {
			m.Write(b[:n])
		}

	}

	si := fs.Size() - ONEPIECE
	if si < 0 {
		si = 0
	}
	f.Seek(si, 0)

	n, e := f.Read(b)
	if e != nil && e != io.EOF {
		return "", e
	}
	if n > 0 {
		m.Write(b[:n])
	}

	return fmt.Sprintf("%x", m.Sum(nil)), nil
}

// "unix\nsocket\nhttp://host:port/uri"
func DownloadFile(ctx context.Context, url, dstpath, tmpath, fileMd5 string, headers map[string]string, headerHook func(http.Header) error, md5mode int) (http.Header, error) {
	var (
		err       error
		n         int64
		h         hash.Hash
		tmpDst    *os.File
		dstWriter *bufio.Writer
	)

	client := http.DefaultClient

	if dstpath != "" {
		if err = os.MkdirAll(path.Dir(dstpath), 0755); err != nil {
			return nil, fmt.Errorf("create dstdir file: %s", err.Error())
		}
	}

	if tmpath != "" {
		if err = os.MkdirAll(path.Dir(tmpath), 0755); err != nil {
			return nil, fmt.Errorf("create tmpdir file: %s", err.Error())
		}

		if tmpDst, err = os.Create(tmpath); err != nil {
			return nil, fmt.Errorf("create tmp file: %s", err.Error())
		}
		dstWriter = bufio.NewWriter(tmpDst)
	}

	// 删除临时文件
	defer func() {
		if dstpath != "" && tmpath != "" {
			tmpDst.Close()
			os.Remove(tmpath)
		}
	}()

	// unix domain socket?
	if strings.HasPrefix(url, "unix") {
		urlfild := strings.Split(url, "\n")
		if len(urlfild) != 3 {
			return nil, errors.New("url err")
		}

		url = urlfild[2]
		client = &http.Client{
			Transport: &http.Transport{
				Dial: func(proto, addr string) (conn net.Conn, err error) {
					return net.Dial("unix", urlfild[1])
				},
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %s", err.Error())
	}

	// header
	if headers != nil {
		for k, v := range headers {
			if k == "Host" {
				request.Host = v
				continue
			}
			request.Header.Set(k, v)
		}
	}

	client.Timeout = 1 * time.Hour

	// request
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http.Do: %s", err.Error())
	}
	defer response.Body.Close()

	// 处理头信息
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.Header, fmt.Errorf("http.StatusCode: %d", response.StatusCode)
	}
	if headerHook != nil {
		if err = headerHook(response.Header); err != nil {
			return response.Header, err
		}
	}

	if "" != fileMd5 {
		h = md5.New()
	}

	// 读Body
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	// download
	for {
		var nr int
		var er error

		select {
		case <-ctx.Done():
			goto CANCEL
		default:
			nr, er = response.Body.Read(buf)
		}

		n = n + int64(nr)
		if nr > 0 && dstWriter != nil {
			nw, ew := dstWriter.Write(buf[:nr])
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}

			// md5
			if "" != fileMd5 && MD5MODEALL == md5mode {
				h.Write(buf[0:nr])
			}
		}
		if er == io.EOF {
			err = nil
			break
		}
		if er != nil {
			err = er
			break
		}
	}
CANCEL:
	if err != nil {
		return response.Header, fmt.Errorf("downloading file %d: %s", n, err.Error())
	}

	// 没有读完
	if n != response.ContentLength {
		response.Header.Set("Readn", fmt.Sprintf("%d", n))
	}

	if dstWriter != nil {
		if err = dstWriter.Flush(); err != nil {
			return response.Header, fmt.Errorf("flush tmp file: %s", err.Error())
		}
	}

	// check md5
	if "" != fileMd5 {
		nmd5 := ""
		if md5mode == MD5MODEALL {
			nmd5 = fmt.Sprintf("%x", h.Sum(nil))
		} else if md5mode == MD5MODE1M {
			if nmd5, err = Md51m(tmpDst); err != nil {
				return response.Header, fmt.Errorf("md5 err: %s", err.Error())
			}
		}

		if fileMd5 != nmd5 {
			return response.Header, fmt.Errorf("md5 err: %s", nmd5)
		}
	}

	if dstpath != "" && tmpath != "" {
		if err = tmpDst.Close(); err != nil {
			return response.Header, fmt.Errorf("close tmp file: %s", err.Error())
		}

		if err = os.Rename(tmpath, dstpath); err != nil {
			return response.Header, fmt.Errorf("rename: %s", err.Error())
		}

		dstStat, err := os.Stat(dstpath)
		if err != nil && os.IsNotExist(err) {
			return response.Header, errors.New("download err")
		}
		if response.ContentLength >= 0 && dstStat.Size() != response.ContentLength {
			return response.Header, fmt.Errorf("size err: %s %d %d", dstpath, response.ContentLength, dstStat.Size())
		}
	}

	return response.Header, nil
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	if UploadDir == "" {
		panic("文件上传存放到哪？")
	}

	if r.Method != "POST" {
		HttpErr(w, http.StatusMethodNotAllowed, -1, "必须是POST")
		return
	}

	r.ParseMultipartForm(32 << 20)

	file, h, err := r.FormFile("file")
	if err != nil {
		log.Println("FileUpload err: ", err)
		HttpErr(w, http.StatusInternalServerError, -1, err.Error())
		return
	}
	defer file.Close()

	fileBuff, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("upload err: ", err)
		HttpErr(w, http.StatusInternalServerError, -1, err.Error())
		return
	}

	fp := fmt.Sprintf("%s%x%s", UploadDir, md5.Sum(fileBuff), path.Ext(h.Filename))

	// 是否已经存在
	if false == IsExists(fp) {
		if ioutil.WriteFile(fp, fileBuff, 0755) != nil {
			log.Println("FileUpload err: ", err)
			HttpErr(w, http.StatusInternalServerError, -1, err.Error())
			return
		}
	}

	log.Println("FileUpload ok: ", fp)

	HttpErr(w, http.StatusOK, 0, fmt.Sprintf("%x%s", md5.Sum(fileBuff), path.Ext(h.Filename)))

	return
}

func PostRequest(url string, body []byte, headers *map[string]string, cookies []*http.Cookie) (response *http.Response, respBody []byte, err error) {
	if url == "" {
		return nil, nil, errors.New("url nil.")
	}

	// body
	bodyBuff := bytes.NewBuffer(body)
	requestReader := io.MultiReader(bodyBuff)
	request, err := http.NewRequest("POST", url, requestReader)
	if err != nil {
		return
	}

	// cookies
	if cookies != nil {
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
	}

	// header
	request.Header.Add("Content-Type", "text/html")
	request.ContentLength = int64(bodyBuff.Len())
	if headers != nil {
		for k, v := range *headers {
			request.Header.Add(k, v)
		}
	}

	// request
	http.DefaultClient.Timeout = 5 * time.Second
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	respBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	return
}

func GetRequest(url string, headers map[string]string) (resp *http.Response, body []byte, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	// header
	if headers != nil {
		for k, v := range headers {
			if k == "Host" {
				request.Host = v
				continue
			}
			request.Header.Set(k, v)
		}
	}

	// request
	http.DefaultClient.Timeout = 1 * time.Minute
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	return response, body, nil
}

func HeadRequest(url string, headers map[string]string) (response *http.Response, err error) {
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	// header
	if headers != nil {
		for k, v := range headers {
			if k == "Host" {
				request.Host = v
				continue
			}
			request.Header.Set(k, v)
		}
	}

	// request
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	response.Body.Close()

	return response, nil
}

func HttpErr(w http.ResponseWriter, statCode int, errno int, message interface{}) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statCode)

	defer w.(http.Flusher).Flush()

	if errno != 0 {
		if _, e := fmt.Fprintf(w, "{\"errcode\":%d,\"errmsg\":\"%s\"}", errno, message); e != nil {
			return e
		}
	} else {
		resp := struct {
			ErrCode int         `json:"errcode"`
			ErrMsg  string      `json:"errmsg"`
			Data    interface{} `json:"data"`
		}{0, "", message}

		b, _ := json.Marshal(resp)
		if _, e := w.Write(b); e != nil {
			return e
		}
	}

	return nil
}
