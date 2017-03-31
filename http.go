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
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024*1024)
	},
}

// "unix socket http://host:port/uri"
func DownloadFile(url, dstpath, tmpath, fileMd5 string, headers map[string]string) (http.Header, error) {
	var (
		err    error
		h      hash.Hash
		dst    *os.File
		client *http.Client = http.DefaultClient
	)

	if err = os.MkdirAll(path.Dir(dstpath), 0755); err != nil {
		return nil, fmt.Errorf("create dst file: %s", err.Error())
	}

	if dst, err = os.Create(tmpath); err != nil {
		return nil, fmt.Errorf("create tmp file: %s", err.Error())
	}
	dstWriter := bufio.NewWriter(dst)

	defer func() { // 删除临时文件
		if dst != nil {
			_ = dst.Close()
		}
		if _, err := os.Stat(tmpath); err == nil || os.IsExist(err) {
			_ = os.Remove(tmpath)
		}
	}()

	if "" != fileMd5 {
		h = md5.New()
	}

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

	// request
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http.Do: %s", err.Error())
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return response.Header, fmt.Errorf("http.StatusCode: %d", response.StatusCode)
	}

	var contentLength int = int(response.ContentLength)
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	// download
	n := 0
	for {
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			n = n + nr
			nw, ew := dstWriter.Write(buf[0:nr])
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}

			if response.ContentLength > 0 {
				contentLength = contentLength - nr
			}

			// md5
			if "" != fileMd5 {
				nh, eh := h.Write(buf[0:nr])
				if eh != nil {
					err = eh
					break
				}
				if nh != nr {
					err = io.ErrShortWrite
					break
				}
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
		return response.Header, fmt.Errorf("downloading file: %s", err.Error())
	}
	if err = dst.Close(); err != nil {
		return response.Header, fmt.Errorf("close tmp file: %s", err.Error())
	}

	if contentLength != 0 && contentLength != -1 {
		return response.Header, fmt.Errorf("short body %v %v", response.ContentLength, n)
	}
	
	if n == 0 {
		return response.Header, fmt.Errorf("zero body %v", response.ContentLength)
	}
	
	// check md5
	if "" != fileMd5 {
		nmd5 := fmt.Sprintf("%x", h.Sum(nil))
		if fileMd5 != nmd5 {
			if e := os.Remove(tmpath); e != nil {
				return response.Header, e
			}

			return response.Header, errors.New("md5 err")
		}
	}

	if err = os.Rename(tmpath, dstpath); err != nil {
		return response.Header, fmt.Errorf("rename: %s", err.Error())
	}

	if _, err = os.Stat(dstpath); err != nil && os.IsNotExist(err) {
		return response.Header, errors.New("download err")
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

func GetRequest(res string, headers *map[string]string) (statusCode int, responseCookies []*http.Cookie, body []byte, err error) {
	if res == "" {
		return 0, nil, nil, errors.New("url nil")
	}

	request, err := http.NewRequest("GET", res, nil)
	if err != nil {
		return 0, nil, nil, err
	}

	// header
	request.Header.Add("Content-Type", "text/html")
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

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, nil, err
	}

	return response.StatusCode, response.Cookies(), body, nil
}

func HttpErr(w http.ResponseWriter, statCode int, message string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statCode)
	if _, e := fmt.Fprintf(w, "{\"message\":\"%s\"}", message); e != nil {
		panic(e)
	}
	w.(http.Flusher).Flush()
}
