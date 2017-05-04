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

// "unix socket http://host:port/uri"
func DownloadFile(url, dstpath, tmpath, fileMd5 string, headers map[string]string, md5mode int) (http.Header, error) {
	var (
		err      error
		n        int64
		h        hash.Hash
		tmpDst   *os.File
		client   *http.Client = http.DefaultClient
		lastonem []byte
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
		if tmpDst != nil {
			_ = tmpDst.Close()
		}
		if _, err := os.Stat(tmpath); err == nil || os.IsExist(err) {
			_ = os.Remove(tmpath)
		}
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
	if response.StatusCode != http.StatusOK {
		return response.Header, fmt.Errorf("http.StatusCode: %d", response.StatusCode)
	}

	var contentLength int64 = response.ContentLength
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	if "" != fileMd5 {
		h = md5.New()

		if MD5MODE1M == md5mode {
			io.WriteString(h, fmt.Sprintf("%d", contentLength)) // 文件长度
		}
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
			if "" != fileMd5 {
				if MD5MODEALL == md5mode {
					h.Write(buf[0:nr])
				} else if MD5MODE1M == md5mode {
					// 最后1M
					if response.ContentLength-n <= ONEMBYTE {
						s := 0

						if response.ContentLength-ONEMBYTE > 0 {
							if n-(response.ContentLength-ONEMBYTE) < int64(nr) {
								s = int(int64(nr) - (n - (response.ContentLength - ONEMBYTE)))
							}
						}

						lastonem = append(lastonem, buf[s:nr]...)
					}

					// 每100M取第1M
					if (n-int64(nr))%(100*ONEMBYTE) >= 0 && (n-int64(nr))%(100*ONEMBYTE) < ONEMBYTE {
						s, e := 0, nr

						if n%(100*ONEMBYTE) < int64(nr) {
							s = nr - int(n%(100*ONEMBYTE))
						}
						if n%(100*ONEMBYTE) >= ONEMBYTE {
							e = e - int(n%(100*ONEMBYTE)-ONEMBYTE)
						}
						h.Write(buf[s:e])
					}
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
	if err = dstWriter.Flush(); err != nil {
		return response.Header, fmt.Errorf("flush tmp file: %s", err.Error())
	}
	if err = tmpDst.Close(); err != nil {
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
		if md5mode == MD5MODE1M {
			h.Write(lastonem[:])
		}

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
