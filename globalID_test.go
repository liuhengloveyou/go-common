package common_test

import (
	"fmt"
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestGlobalID(t *testing.T) {
	g := &common.GlobalID{Type: "T", ServID: "S", Expand: "E"}

	for i := 0; i < 100; i++ {
		fmt.Println(g.ID())
	}
}
