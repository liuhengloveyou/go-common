package common_test

import (
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestAes(t *testing.T) {
	text := "abcdefg"
	key := "aaaaaaaaaaaaaaaasadfasdfasdf"

	str, err := common.AesCBCEncrypt(text, key)
	t.Error(str, err)

	rst, err := common.AesCBCDecrypt(str, key)
	t.Error(rst, err)
}
