package password

import (
	"encoding/base64"
	"io"
	"math/rand"
	"net"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func RandPassword() []byte {
	intArr := rand.Perm(256)
	var password = make([]byte, 256)
	for i, v := range intArr {
		password[i] = byte(v)
		if i == v {
			return RandPassword()
		}
	}
	return password
}

func ParsePassword(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

type PasswordRW struct {
	conn           net.Conn
	encodePassword []byte
	decodePassword []byte
}

func (p2 *PasswordRW) Read(p []byte) (n int, err error) {
	var b = make([]byte, len(p))
	n, err = p2.conn.Read(b)
	if err != nil {
		return 0, err
	}
	var x = make([]byte, n)
	for i, v := range b[:n] {
		x[i] = p2.decodePassword[v]
	}
	copy(p[:n], x[:n])
	return
}

func (p2 *PasswordRW) Write(p []byte) (n int, err error) {
	var x = make([]byte, len(p))
	for i, v := range p {
		x[i] = p2.encodePassword[v]
	}
	return p2.conn.Write(x)
}

func NewPasswordRW(password []byte, conn net.Conn) io.ReadWriter {
	ppw := new(PasswordRW)
	ppw.conn = conn
	ppw.encodePassword = make([]byte, 256)
	ppw.decodePassword = make([]byte, 256)
	for i, v := range password {
		ppw.decodePassword[v] = uint8(i)
	}
	ppw.encodePassword = password
	return ppw
}
