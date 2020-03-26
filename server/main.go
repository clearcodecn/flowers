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
			fmt.Println("password is: ", string(x))
			ioutil.WriteFile(file, []byte(x), 0777)
		}
	} else {
		data, _ := ioutil.ReadFile(file)
		fmt.Println("password is: ", string(data))
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

// TODO:: keep alive with dst.
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
	//	header长度， uint16 解码得出domain的长度
	//	-----------------
	//	   2  |   domain
	//	------------------
	l := binary.BigEndian.Uint16(b[:n])
	b = make([]byte, l)
	n, err = pcc.Read(b)
	if err != nil {
		log.Println(err)
		return
	}
	// 得到remote的host和port，建立连接
	hostPort := string(b[:n])
	dst, err := net.Dial("tcp", hostPort)
	if err != nil {
		log.Println(err)
		return
	}
	defer dst.Close()

	log.Println("proxy", hostPort)
	// 然后 pipe 数据
	go io.Copy(dst, pcc)
	io.Copy(pcc, dst)
}
