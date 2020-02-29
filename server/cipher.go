package server

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

type Cipher interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type DictCipher struct {
	password [256]byte

	key []byte
}

func (d *DictCipher) Encode(dst []byte) ([]byte, error) {
	return AesEncrypt(dst, d.key)
}

func (d *DictCipher) Decode(src []byte) ([]byte, error) {
	return AesDecrypt(src, d.key)
}

func NewDictCipher(k []byte) (Cipher, error) {
	dc := new(DictCipher)
	m := md5.New()
	m.Write(k)
	x := m.Sum(nil)
	key := fmt.Sprintf("%x", x)
	dc.key = []byte(key)
	return dc, nil
}

func AesEncrypt(orig []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	orig = PKCS7Padding(orig, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	dst := make([]byte, len(orig))

	blockMode.CryptBlocks(dst, orig)
	return []byte(base64.StdEncoding.EncodeToString(dst)), nil
}
func AesDecrypt(dst []byte, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(string(dst))
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	orig := make([]byte, len(data))
	blockMode.CryptBlocks(orig, data)
	orig = PKCS7UnPadding(orig)
	return orig, nil
}

func PKCS7Padding(text []byte, blockSize int) []byte {
	padding := blockSize - len(text)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(text, padText...)
}

func PKCS7UnPadding(origin []byte) []byte {
	length := len(origin)
	unPadding := int(origin[length-1])
	return origin[:(length - unPadding)]
}
