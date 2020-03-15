package main

import (
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/clearcodecn/flowers/password"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
)

var (
	addr string
	pswd []byte
)

func init() {
	flag.StringVar(&addr, "a", ":9898", "address")
}

func main() {
	flag.Parse()

	p, _ := os.UserHomeDir()

	file := filepath.Join(p, "iproxy")
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			pswd = password.RandPassword()
			x := base64.StdEncoding.EncodeToString(pswd)
			fmt.Println("密码: ", string(x))
			ioutil.WriteFile(file, []byte(x), 0777)
		}
	} else {
		data, _ := ioutil.ReadFile(file)
		fmt.Println("密码: ", string(data))
		pswd, _ = base64.StdEncoding.DecodeString(string(data))
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("listen: ", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	pcc := password.NewPasswordRW(pswd, conn)

	// read host & port
	var b = make([]byte, 2)
	n, err := pcc.Read(b)
	if err != nil {
		log.Println(err)
		return
	}
	l := binary.BigEndian.Uint16(b[:n])
	b = make([]byte, l)
	n, err = pcc.Read(b)
	if err != nil {
		log.Println(err)
		return
	}

	hostPort := string(b[:n])

	dst, err := net.Dial("tcp", hostPort)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("proxy", hostPort)

	go io.Copy(dst, pcc)
	io.Copy(pcc, dst)
}