package common_test

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestGetRequest(t *testing.T) {
	h, c, b, e := common.GetRequest("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jp", nil)
	fmt.Println(h, c, b, e)
}

func TestDownload(t *testing.T) {

	m := md5.New()
	resp, _ := http.Get("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg")
	io.WriteString(m, fmt.Sprintf("%d", resp.ContentLength)) // 文件长度
	body, _ := ioutil.ReadAll(resp.Body)
	m.Write(body) // 第1m
	m.Write(body) // 最后1m
	fmd5 := fmt.Sprintf("%x", m.Sum(nil))
	fmt.Println(">>>", resp.ContentLength, fmd5)

	wr, err := common.DownloadFile("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg",
		"/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", fmd5, nil, common.MD5MODE1M)
	fmt.Printf("%#v; %v\n\n", wr, err)

	/*
		m := md5.New()

		request, _ := http.NewRequest("GET", "http://121.12.98.27:80/tpr/wow/data/89/00/890046363a7f8ea4a97db98daed02765", nil)
		request.Host = "client03.pdl.wow.battlenet.com.cn"
		resp, _ := http.DefaultClient.Do(request)
		body, _ := ioutil.ReadAll(resp.Body)
		io.WriteString(m, fmt.Sprintf("%d", resp.ContentLength))           // 文件长度
		m.Write(body[0 : 1024*1024])                                       // 第1m
		m.Write(body[100*common.ONEMBYTE : 100*common.ONEMBYTE+1024*1024]) // 第2m
		m.Write(body[200*common.ONEMBYTE : 200*common.ONEMBYTE+1024*1024]) // 第3m
		m.Write(body[resp.ContentLength-1024*1024 : resp.ContentLength])   // 最后1m
		fmd5 := fmt.Sprintf("%x", m.Sum(nil))
		fmt.Println(resp.ContentLength, fmd5)

		wr, err := common.DownloadFile("http://127.0.0.1:8080/download/fff",
			"/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", fmd5, nil, common.MD5MODE1M)
		fmt.Printf("%#v; %v\n\n", wr, err)
	*/
}
