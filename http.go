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
	"net/http/httptrace"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// 文件下载
const ONEPIECE = 10 * 1024
const (
	MD5MODEALL = iota
	MD5MODE1M  /*文件长度+每隔1M取开始10K+文件尾10K*/
)

type HttpErrMsg struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"errmsg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type Downloader struct {
	Headers    map[string]string
	HeaderHook func(http.Header) error

	URL     string
	DstPath string
	TmPath  string

	MD5     string
	MD5mode int

	Trace *httptrace.ClientTrace

	// 间隔多少秒打印下载日志; 0表示结束时打; -1不打
	LogMode int64
	LogHook func(nn, n int64)

	// 下载了多少字节
	N, nn int64
}

// 上传接口把文件放在哪个目录
var UploadDir string

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
func (p *Downloader) Download(ctx context.Context) (resp *http.Response, err error) {
	var (
		lastLogTime int64
		h           hash.Hash
		tmpDst      *os.File
		dstWriter   *bufio.Writer
	)

	client := http.DefaultClient

	if p.TmPath != "" {
		if err = os.MkdirAll(path.Dir(p.TmPath), 0755); err != nil {
			return nil, fmt.Errorf("create tmpdir file: %s", err.Error())
		}

		if tmpDst, err = os.Create(p.TmPath); err != nil {
			return nil, fmt.Errorf("create tmp file: %s", err.Error())
		}
		dstWriter = bufio.NewWriter(tmpDst)
	}

	if p.DstPath != "" {
		if err = os.MkdirAll(path.Dir(p.DstPath), 0755); err != nil {
			return nil, fmt.Errorf("create dstdir file: %s", err.Error())
		}

		if dstWriter == nil {
			if tmpDst, err = os.Create(p.DstPath); err != nil {
				return nil, fmt.Errorf("create dst file: %s", err.Error())
			}
			dstWriter = bufio.NewWriter(tmpDst)
		}
	}

	// 删除临时文件
	defer func() {
		if p.DstPath != "" && p.TmPath != "" {
			tmpDst.Close()
			os.Remove(p.TmPath)
		}
	}()

	// unix domain socket?
	if strings.HasPrefix(p.URL, "unix") {
		urlfild := strings.Split(p.URL, "\n")
		if len(urlfild) != 3 {
			return nil, errors.New("url err")
		}

		p.URL = urlfild[2]
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

	request, err := http.NewRequest("GET", p.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %s", err.Error())
	}
	request = request.WithContext(ctx)
	if p.Trace != nil {
		request = request.WithContext(httptrace.WithClientTrace(request.Context(), p.Trace))
	}

	// header
	if p.Headers != nil {
		for k, v := range p.Headers {
			if k == "Host" {
				request.Host = v
				continue
			}
			request.Header.Set(k, v)
		}
	}

	client.Timeout = 1 * time.Hour

	// request
	var response *http.Response
	if response, err = client.Do(request); err != nil {
		return nil, fmt.Errorf("http.Do: %s", err.Error())
	}
	defer response.Body.Close()

	// 处理头信息
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response, fmt.Errorf("http.StatusCode: %d", response.StatusCode)
	}

	if p.HeaderHook != nil {
		if err = p.HeaderHook(response.Header); err != nil {
			return response, err
		}
	}

	if "" != p.MD5 {
		h = md5.New()
	}

	// 读Body
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	// download
	for {
		nr, er := response.Body.Read(buf)
		p.N, p.nn = p.N+int64(nr), p.nn+int64(nr)

		nt := time.Now().Unix()
		if p.LogHook != nil && ((p.LogMode > 0 && nt%p.LogMode == 0 && nt != lastLogTime) || er != nil) {
			p.LogHook(p.nn, p.N)
			p.nn, lastLogTime = 0, nt
		}

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
			if p.MD5 != "" && MD5MODEALL == p.MD5mode {
				h.Write(buf[0:nr])
			}
		}

		if er != nil {
			err = er
			break
		}
	}

	if err != nil && err != io.EOF && err != context.DeadlineExceeded && err != context.Canceled {
		return response, fmt.Errorf("downloading file %d: %s", p.N, err.Error())
	}

	if dstWriter != nil {
		if err = dstWriter.Flush(); err != nil {
			return response, fmt.Errorf("flush tmp file: %s", err.Error())
		}
	}

	// check md5
	if "" != p.MD5 {
		nmd5 := ""
		if p.MD5mode == MD5MODEALL {
			nmd5 = fmt.Sprintf("%x", h.Sum(nil))
		} else if p.MD5mode == MD5MODE1M {
			if nmd5, err = Md51m(tmpDst); err != nil {
				return response, fmt.Errorf("md5 err: %s", err.Error())
			}
		}

		if p.MD5 != nmd5 {
			return response, fmt.Errorf("md5 err: %s", nmd5)
		}
	}

	if p.DstPath != "" && p.TmPath != "" {
		if err = tmpDst.Close(); err != nil {
			return response, fmt.Errorf("close tmp file: %s", err.Error())
		}

		if err = os.Rename(p.TmPath, p.DstPath); err != nil {
			return response, fmt.Errorf("rename: %s", err.Error())
		}

		dstStat, err := os.Stat(p.DstPath)
		if err != nil && os.IsNotExist(err) {
			return response, errors.New("download err")
		}
		if response.ContentLength >= 0 && dstStat.Size() != response.ContentLength {
			return response, fmt.Errorf("size err: %s %d %d", p.DstPath, response.ContentLength, dstStat.Size())
		}
	}

	return response, nil
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	if UploadDir == "" {
		panic("文件上传存放到哪？")
	}

	if !strings.HasSuffix(UploadDir, "/") {
		UploadDir = UploadDir + "/"
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

func PostRequest(url string, headers map[string]string, cookies []*http.Cookie, body ...[]byte) (response *http.Response, respBody []byte, err error) {
	if url == "" {
		return nil, nil, errors.New("url nil.")
	}

	// body
	contentLength := 0
	readers := make([]io.Reader, len(body))
	for i := 0; i < len(body); i++ {
		readers[i] = bytes.NewBuffer(body[i])
		contentLength = contentLength + len(body[i])
	}
	requestReader := io.MultiReader(readers...)
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
	request.ContentLength = int64(contentLength)
	if len(headers) > 0 {
		for k, v := range headers {
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
	http.DefaultClient.Timeout = 100 * time.Second
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, nil, err
		}
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

func HttpErr(w http.ResponseWriter, code, errno int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	defer w.(http.Flusher).Flush()

	errMsg := HttpErrMsg{Code: errno}
	if errno != 0 {
		errMsg.Msg = data
	} else {
		errMsg.Data = data
	}

	b, _ := json.Marshal(errMsg)
	w.Write(b)

	return
}

func UnmarshalHttpResponse(data []byte, v interface{}) error {
	rst := HttpErrMsg{
		Data: v,
	}

	return json.Unmarshal(data, &rst)
}