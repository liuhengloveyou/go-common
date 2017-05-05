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

	///////////////////
	m := md5.New()
	resp, _ := http.Get("https://d11.baidupcs.com/file/a5ad93188fcd4ea954ff5e9591a0cd10?bkt=p3-1400a5ad93188fcd4ea954ff5e9591a0cd10c5e256bb000000118a6f&xcode=ba0a287c6e124c3f1891970f8cd464512474ad6a75f04f0f9717ec4418c70769&fid=1983562431-250528-537546744507152&time=1493878729&sign=FDTAXGERLBHS-DCb740ccc5511e5e8fedcff06b081203-pU7PgG6NAFYYjROUaunhVBJNvKA%3D&to=d11&size=1149551&sta_dx=1149551&sta_cs=35407&sta_ft=mobi&sta_ct=6&sta_mt=6&fm2=MH,Yangquan,Netizen-anywhere,,beijing,any&newver=1&newfm=1&secfm=1&flow_ver=3&pkey=1400a5ad93188fcd4ea954ff5e9591a0cd10c5e256bb000000118a6f&sl=83034191&expires=8h&rt=sh&r=880065485&mlogid=2866549678943103904&vuk=1983562431&vbdid=2754292345&fin=%E4%BB%8E0%E5%88%B01.mobi&fn=%E4%BB%8E0%E5%88%B01.mobi&rtype=1&iv=0&dp-logid=2866549678943103904&dp-callid=0.1.1&hps=1&csl=300&csign=0ClNecZSMgFr%2BWHYnSP9KuFBcxo%3D&by=themis")
	body, _ := ioutil.ReadAll(resp.Body)
	io.WriteString(m, fmt.Sprintf("%d", resp.ContentLength))         // 文件长度
	m.Write(body[0 : 1024*1024])                                     // 第1m
	m.Write(body[resp.ContentLength-1024*1024 : resp.ContentLength]) // 最后1m
	fmd5 := fmt.Sprintf("%x", m.Sum(nil))
	fmt.Println(resp.ContentLength, fmd5)

	wr, err := common.DownloadFile("https://d11.baidupcs.com/file/a5ad93188fcd4ea954ff5e9591a0cd10?bkt=p3-1400a5ad93188fcd4ea954ff5e9591a0cd10c5e256bb000000118a6f&xcode=ba0a287c6e124c3f1891970f8cd464512474ad6a75f04f0f9717ec4418c70769&fid=1983562431-250528-537546744507152&time=1493878729&sign=FDTAXGERLBHS-DCb740ccc5511e5e8fedcff06b081203-pU7PgG6NAFYYjROUaunhVBJNvKA%3D&to=d11&size=1149551&sta_dx=1149551&sta_cs=35407&sta_ft=mobi&sta_ct=6&sta_mt=6&fm2=MH,Yangquan,Netizen-anywhere,,beijing,any&newver=1&newfm=1&secfm=1&flow_ver=3&pkey=1400a5ad93188fcd4ea954ff5e9591a0cd10c5e256bb000000118a6f&sl=83034191&expires=8h&rt=sh&r=880065485&mlogid=2866549678943103904&vuk=1983562431&vbdid=2754292345&fin=%E4%BB%8E0%E5%88%B01.mobi&fn=%E4%BB%8E0%E5%88%B01.mobi&rtype=1&iv=0&dp-logid=2866549678943103904&dp-callid=0.1.1&hps=1&csl=300&csign=0ClNecZSMgFr%2BWHYnSP9KuFBcxo%3D&by=themis",
		"/tmp/a/b/c/aaa.jpg", "/tmp/bbb/ccc/ddd", fmd5, nil, common.MD5MODE1M)
	fmt.Printf("%#v; %v\n\n", wr, err)

	/*
		////////////
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
