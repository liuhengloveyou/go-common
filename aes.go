package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func getKey(key string) []byte {
	for len(key) < 32 {
		key = key + key
	}

	return []byte(key)[:32]
}

//加密字符串
func AesCBCEncrypt(msg, key string) (string, error) {
	plantText := []byte(msg)
	rkey := getKey(key)

	block, err := aes.NewCipher(rkey) //选择加密算法
	if err != nil {
		return "", err
	}

	plantText = PKCS7Padding(plantText, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, rkey[:aes.BlockSize])

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)

	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

//解密字符串
func AesCBCDecrypt(msg, key string) (rst string, err error) {
	defer func() {
		//错误处理
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	rmsg, err := base64.RawStdEncoding.DecodeString(msg)
	if err != nil {
		return "", err
	}

	rkey := getKey(key)
	block, err := aes.NewCipher(rkey)
	if err != nil {
		return "", err
	}

	blockModel := cipher.NewCBCDecrypter(block, rkey[:aes.BlockSize])
	plantText := make([]byte, len(rmsg))
	blockModel.CryptBlocks(plantText, rmsg)
	plantText = PKCS7UnPadding(plantText, block.BlockSize())

	return string(plantText), nil
}

func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
