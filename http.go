package common

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func DownloadFile(url, path, tmpath, fileMd5 string, headers map[string]string) (http.Header, error) {
	var (
		err error
		h   hash.Hash
		dst *os.File
	)

	if dst, err = os.Create(tmpath); err != nil {
		return nil, err
	}

	if "" != fileMd5 {
		h = md5.New()
	}

	buf := make([]byte, 32*1024)

	request, err := http.NewRequest("GET", url, nil)
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
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d", response.StatusCode)
	}

	// download
	for {
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
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
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	if err = dst.Close(); err != nil {
		return nil, err
	}

	// check md5
	if "" != fileMd5 {
		nmd5 := fmt.Sprintf("%x", h.Sum(nil))
		if fileMd5 != nmd5 {
			if e := os.Remove(tmpath); e != nil {
				return nil, e
			}

			return nil, fmt.Errorf("md5 ERR: ", fileMd5, nmd5)
		}
	}

	if err = os.Rename(tmpath, path); err != nil {
		return nil, err
	}

	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("download err.")
		}
	}

	return response.Header, nil
}

func PostRequest(path string, body []byte, headers *map[string]string, cookies []*http.Cookie) (statuCode int, responseCookies []*http.Cookie, responseBody []byte, err error) {
	if path == "" {
		return 0, nil, nil, fmt.Errorf("path nil.")
	}

	// body
	bodyBuff := bytes.NewBuffer(body)
	requestReader := io.MultiReader(bodyBuff)
	request, err := http.NewRequest("POST", path, requestReader)
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

func GetRequest(path string, headers *map[string]string) (statusCode int, responseCookies []*http.Cookie, body []byte, err error) {
	if path == "" {
		return 0, nil, nil, fmt.Errorf("URL nil")
	}

	request, err := http.NewRequest("GET", path, nil)
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
