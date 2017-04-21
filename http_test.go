package common_test

import (
	"fmt"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestDownload(t *testing.T) {
	wr, err := common.DownloadFile("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg", "/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", "", nil)
	fmt.Printf("%#v; %v\n\n", wr, err)

	wr, err = common.DownloadFile("unix /tmp/nginx.sock http://127.0.0.1:8080/lua", "/tmp/a/b/c/lua", "/tmp/bbb/bbb", "", nil)
	fmt.Printf("%#v; %v\n\n", wr, err)

}
