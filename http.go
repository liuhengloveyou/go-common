package common

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func PostRequest(path string, body []byte, headers *map[string]string) (status int, cookies []*http.Cookie, responseBody []byte, err error) {
	if path == "" {
		return 0, nil, nil, fmt.Errorf("path nil.")
	}
	if body == nil {
		return 0, nil, nil, fmt.Errorf("body nil.")
	}

	// body
	bodyBuff := bytes.NewBuffer(body)
	requestReader := io.MultiReader(bodyBuff)
	request, err := http.NewRequest("POST", path, requestReader)
	if err != nil {
		return 0, nil, nil, err
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

	status = response.StatusCode
	cookies = response.Cookies()
	responseBody, err = ioutil.ReadAll(response.Body)

	return
}

func GetRequest(path string) (statusCode int, body []byte, err error) {
	if path == "" {
		return 0, nil, fmt.Errorf("URL nil")
	}

	response, err := http.Get(path)
	if err != nil {
		return 0, nil, err
	}

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}
	response.Body.Close()

	return response.StatusCode, body, nil
}

func HttpErr(w http.ResponseWriter, statCode int, body []byte) {
	w.WriteHeader(statCode)
	w.Write(body)
}