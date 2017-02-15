package common_test

import (
	"fmt"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestDownload(t *testing.T) {
	url := "http://b.aa.cm/image/77c6a7efce1b9d1634356c61f1deb48f8d5464c4.jpg"
	wr, err := common.DownloadFile(url, "/tmp/aaa.jpg", "/tmp/bbb", "aacbfda14fee114e7849738bd9e68623", nil)
	fmt.Printf("%#v; %v", wr, err)
}
