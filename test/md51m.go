package main

import (
	"fmt"
	"os"
	"time"

	common "github.com/liuhengloveyou/go-common"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "<文件名>")
		return
	}

	f, e := os.Open(os.Args[1])
	if e != nil {
		panic(e)
	}

	fmd5, e := common.Md51m(f)
	if e != nil {
		panic(e)
	}
	fmt.Printf("%v %v", time.Now().Unix(), fmd5)

	return
}
