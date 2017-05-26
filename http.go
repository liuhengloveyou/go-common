package common

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const ONEMBYTE = 1 * 1024 * 1024

const (
	MD5MODEALL = iota
	MD5MODE1M  /*文件长度+每隔100M取开始1M+文件尾1M*/
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64*1024)
	},
}

func md51m(f *os.File) (string, error) {
	m := md5.New()
	b := make([]byte, ONEMBYTE)

	fs, err := f.Stat()
	if err != nil {
		return "", err
	}

	io.WriteString(m, fmt.Sprintf("%d", fs.Size()))

	var i int64
	for ; i <= (fs.Size() / (100 * ONEMBYTE)); i++ {
		f.Seek(i*100*ONEMBYTE, 0)
		n, _ := f.Read(b)
		m.Write(b[:n])
	}

	si := fs.Size() - ONEMBYTE
	if si < 0 {
		si = 0
	}
	f.Seek(si, 0)

	n, _ := f.Read(b)
	m.Write(b[0:n])

	return fmt.Sprintf("%x", m.Sum(nil)), nil
}

// "unix socket http://host:port/uri"
func DownloadFile(url, dstpath, tmpath, fileMd5 string, headers map[string]string, headerHook func(http.Header) error, md5mode int) (http.Header, error) {
	var (
		err    error
		n      int64
		h      hash.Hash
		tmpDst *os.File
		client *http.Client = http.DefaultClient
	)

	if err = os.MkdirAll(path.Dir(dstpath), 0755); err != nil {
		return nil, fmt.Errorf("create dstdir file: %s", err.Error())
	}

	if err = os.MkdirAll(path.Dir(tmpath), 0755); err != nil {
		return nil, fmt.Errorf("create tmpdir file: %s", err.Error())
	}

	if tmpDst, err = os.Create(tmpath); err != nil {
		return nil, fmt.Errorf("create tmp file: %s", err.Error())
	}
	dstWriter := bufio.NewWriter(tmpDst)

	defer func() { // 删除临时文件
		tmpDst.Close()
		os.Remove(tmpath)
	}()

	// unix domain socket?
	if strings.HasPrefix(url, "unix") {
		urlfild := strings.Fields(url)
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
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.Header, fmt.Errorf("http.StatusCode: %d", response.StatusCode)
	}

	var contentLength int64 = response.ContentLength
	if contentLength < 0 {
		return response.Header, errors.New("unknown contentlength")
	}

	if headerHook != nil {
		if err = headerHook(response.Header); err != nil {
			return response.Header, err
		}
	}

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	if "" != fileMd5 {
		h = md5.New()
	}

	// download
	for {
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			n = n + int64(nr)
			nw, ew := dstWriter.Write(buf[:nr])
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}

			if response.ContentLength > 0 {
				contentLength = contentLength - int64(nr)
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
	if err != nil {
		return response.Header, fmt.Errorf("downloading file %d: %s", n, err.Error())
	}
	if err = dstWriter.Flush(); err != nil {
		return response.Header, fmt.Errorf("flush tmp file: %s", err.Error())
	}

	if contentLength != 0 && contentLength != -1 {
		return response.Header, fmt.Errorf("short body %v %v", response.ContentLength, n)
	}

	if n == 0 {
		return response.Header, fmt.Errorf("zero body %v", response.ContentLength)
	}
	if n != response.ContentLength {
		return response.Header, fmt.Errorf("read %d", n)
	}

	// check md5
	if "" != fileMd5 {
		nmd5 := ""
		if md5mode == MD5MODEALL {
			nmd5 = fmt.Sprintf("%x", h.Sum(nil))
		} else if md5mode == MD5MODE1M {
			if nmd5, err = md51m(tmpDst); err != nil {
				return response.Header, fmt.Errorf("md5 err: %s", err.Error())
			}
		}

		if fileMd5 != nmd5 {
			return response.Header, fmt.Errorf("md5 err: %s", nmd5)
		}
	}

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

	return response.Header, nil
}

func PostRequest(res string, body []byte, headers *map[string]string, cookies []*http.Cookie) (statuCode int, responseCookies []*http.Cookie, responseBody []byte, err error) {
	if res == "" {
		return 0, nil, nil, errors.New("res nil.")
	}

	// body
	bodyBuff := bytes.NewBuffer(body)
	requestReader := io.MultiReader(bodyBuff)
	request, err := http.NewRequest("POST", res, requestReader)
	if err != nil {
		return 0, nil, nil, err
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
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, nil, nil, err
	}
	defer response.Body.Close()

	statuCode = response.StatusCode
	responseCookies = response.Cookies()
	responseBody, err = ioutil.ReadAll(response.Body)

	return
}

func GetRequest(url string, headers map[string]string) (header http.Header, responseCookies []*http.Cookie, body []byte, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, nil, err
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
	http.DefaultClient.Timeout = 1 * time.Hour
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, nil, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	return response.Header, response.Cookies(), body, nil
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

func HttpErr(w http.ResponseWriter, statCode int, message string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statCode)
	if _, e := fmt.Fprintf(w, "{\"message\":\"%s\"}", message); e != nil {
		panic(e)
	}
	w.(http.Flusher).Flush()
}
