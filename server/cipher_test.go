package server

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func makeKey(k string) []byte {
	m := md5.New()
	m.Write([]byte(k))
	return m.Sum(nil)
}

func TestNewDictCipher(t *testing.T) {
	dist, err := NewDictCipher([]byte("abcdefg"))
	if err != nil {
		t.Fatal(err)
	}

	enc, err := dist.Encode([]byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("enc %s", enc)

	b, err := dist.Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}
