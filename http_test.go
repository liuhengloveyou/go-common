package common_test

import (
	"fmt"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestDownload(t *testing.T) {
	url := "http://b.hiphotos.baidu.com/image/pic/item/77c6a7efce1b9d1634356c61f1deb48f8d5464c4.jpg"
	wr, err := common.DownloadFile(url, "/tmp/aaa.jpg", "/tmp/bbb", "", nil)
	fmt.Println(wr, err)
}