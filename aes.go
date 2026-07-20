package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func getKey(key string) []byte {
	for len(key) < 32 {
		key = key + key
	}

	return []byte(key)[:32]
}

//加密字符串
func AesCBCEncrypt(msg []byte, key string, iv []byte) (string, error) {
	b, e := AesCBCEncryptByByte(msg, getKey(key), iv)

	return base64.RawStdEncoding.EncodeToString(b), e
}

func AesCBCEncryptByByte(msg, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}

	if len(iv) == 0 {
		iv = key[:aes.BlockSize]
	}

	msg = PKCS7Padding(msg, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(msg))
	blockModel.CryptBlocks(ciphertext, msg)

	return ciphertext, nil
}

//解密字符串
func AesCBCDecrypt(msg []byte, key string, iv []byte) (rst []byte, err error) {
	return AesCBCDecryptByByte(msg, getKey(key), iv)
}

func AesCBCDecryptByByte(msg, key, iv []byte) (rst []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) == 0 {
		iv = key[:aes.BlockSize]
	}

	blockModel := cipher.NewCBCDecrypter(block, iv)
	plantText := make([]byte, len(msg))
	blockModel.CryptBlocks(plantText, msg)
	plantText = PKCS7UnPadding(plantText)

	return plantText, nil
}

func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
