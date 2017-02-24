package common_test

import (
	"fmt"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestDownload(t *testing.T) {
	url := "http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg"
	wr, err := common.DownloadFile(url, "/tmp/aaa.jpg", "/tmp/bbb", "", nil)
	fmt.Printf("%#v; %v\n\n", wr, err)

	url = "unix /tmp/nginx.sock http://demo:80/lua"
	wr, err = common.DownloadFile(url, "/tmp/lua", "/tmp/bbb", "", nil)
	fmt.Printf("%#v; %v", wr, err)
}
