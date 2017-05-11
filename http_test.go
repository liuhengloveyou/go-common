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
	/*
		m := md5.New()
		resp, _ := http.Get("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg")
		body, _ := ioutil.ReadAll(resp.Body)
		io.WriteString(m, fmt.Sprintf("%d", resp.ContentLength)) // 文件长度
		m.Write(body)                                            // 第1m
		m.Write(body)                                            // 最后1m
		fmd5 := fmt.Sprintf("%x", m.Sum(nil))

		wr, err := common.DownloadFile("http://e.hiphotos.baidu.com/image/pic/item/a1ec08fa513d2697e54749d557fbb2fb4216d8a6.jpg",
			"/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", fmd5, nil, common.MD5MODE1M)
		fmt.Printf("%#v; %v\n\n", wr, err)
	*/

	m := md5.New()

	request, _ := http.NewRequest("GET", "http://127.0.0.1:8080/download/fff", nil)
	resp, err := http.DefaultClient.Do(request)
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

	/*
		m1 := md5.New()
		resp1, _ := http.Get("http://yinyueshiting.baidu.com/data2/music/134380685/620023111600128.mp3?xcode=0772ef241f5ab1fa23407a79428c3f0f")
		body1, _ := ioutil.ReadAll(resp1.Body)
		io.WriteString(m1, fmt.Sprintf("%d", resp1.ContentLength))           // 文件长度
		m1.Write(body1[0 : 1024*1024])                                       // 第1m
		m1.Write(body1[resp1.ContentLength-1024*1024 : resp1.ContentLength]) // 最后1m
		fmd51 := fmt.Sprintf("%x", m1.Sum(nil))

		wr, err = common.DownloadFile("http://yinyueshiting.baidu.com/data2/music/134380685/620023111600128.mp3?xcode=0772ef241f5ab1fa23407a79428c3f0f",
			"/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", fmd51, nil, common.MD5MODE1M)
		fmt.Printf("%#v; %v\n\n", wr, err)
	*/
}
