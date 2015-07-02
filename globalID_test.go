package common_test

import (
	"fmt"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestGlobalID(t *testing.T) {
	g := &common.GlobalID{}
	g.Init("id", "")

	for i := 0; i < 100; i++ {
		fmt.Println(<-g.Hole)
	}
}
