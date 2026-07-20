package common_test

import (
	"testing"

	common "github.com/liuhengloveyou/go-common"
)

func TestAes(t *testing.T) {
	text := "ojQXu0MEwBMIGlscJkR4Qd_9Eurw"
	key := "这是瞧一瞧宝宝的用户ID生成盐,这样比较安全。一点也不能改"

	str, err := common.AesCBCEncrypt([]byte(text), key, nil)
	t.Error(str, err)

	rst, err := common.AesCBCDecrypt([]byte(str), key, nil)
	t.Error(rst, err)
}
