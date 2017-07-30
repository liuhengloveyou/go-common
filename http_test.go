package common_test

import (
	"fmt"
	"net/http"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestGetRequest(t *testing.T) {
	h, b, e := common.GetRequest("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jp", nil)
	fmt.Println(h, b, e)
}

func TestDownload(t *testing.T) {
	wr, err := common.DownloadFile("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg",
		"/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", "c78225bfb0a5eba6490b77f4eabdfc35", nil, func(header http.Header) error { fmt.Println("hook: ", header); return nil }, common.MD5MODE1M)
	fmt.Printf("%#v; %v\n\n", wr, err)
}
