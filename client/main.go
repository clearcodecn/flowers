package main

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/clearcodecn/flowers/ad"
	"github.com/clearcodecn/flowers/password"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var (
	pswd  string
	saddr string
	laddr string

	passAd bool

	p []byte
)

var filterDomain = func(host string) bool {
	return false
}

func init() {
	flag.StringVar(&pswd, "p", "", "password")
	flag.StringVar(&saddr, "s", ":9898", "server address")
	flag.StringVar(&laddr, "l", ":1080", "local address")
	flag.BoolVar(&passAd, "ad", false, "pass ad")
}

func main() {
	flag.Parse()

	if passAd {
		filterDomain = ad.FilterAdDomain()
	}

	p, _ = base64.StdEncoding.DecodeString(pswd)

	if len(p) == 0 {
		fmt.Println("password can not be empty")
		return
	}
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("client proxy running at:", laddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	var b = make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		log.Println("read local connection failed:", err)
		return
	}
	var (
		method, host string
	)
	// 尝试解析 socks5 协议，如果不是，那么就用http(s)协议
	host, err = handleSS5(b[:n], conn)
	if err == nil {
		b = b[:]
		n, err = conn.Read(b)
		if err != nil {
			return
		}
	} else {
		method, host, err = FindHost(b[:n])
		if err != nil {
			log.Println("parse hostPort failed:", err)
			return
		}
		// 如果是 https 协议，给本地连接发送一个已经连接的初始包
		if method == http.MethodConnect {
			_, _ = fmt.Fprint(conn, "HTTP/1.1 200 Connection established\r\n\r\n")
		}
	}

	if passAd {
		h := strings.Split(host, ":")
		if filterDomain(h[0]) {
			conn.Close()

			log.Println("block domain host: ", host)
			return
		}
	}

	l := len(host)
	var pkg = make([]byte, 2)
	binary.BigEndian.PutUint16(pkg, uint16(l))
	pkg = append(pkg, host...)

	// 连接服务器
	dst, err := net.Dial("tcp", saddr)
	if err != nil {
		log.Print(err)
		return
	}

	defer dst.Close()

	log.Println("proxy: ", host)
	cc := password.NewPasswordRW(p, dst)
	cc.Write(pkg)

	// 发送剩余的数据
	if method != http.MethodConnect {
		cc.Write(b[:n])
	}
	// 开始传输数据
	go io.Copy(cc, conn)
	io.Copy(conn, cc)
}

//FindHost
func FindHost(data []byte) (method string, host string, err error) {
	arr := strings.Split(string(data), "\r\n")
	part := arr[0]
	partArr := strings.Split(part, " ")
	method = partArr[0]
	if len(partArr) < 2 {
		return "", "", errors.New("can not parse host:port in first header")
	}
	for _, v := range arr[1:] {
		if strings.HasPrefix(v, "Host: ") {
			host := strings.TrimPrefix(v, "Host: ")
			if strings.Contains(host, ":") {
				return method, host, nil
			}
			if method == http.MethodConnect {
				return method, fmt.Sprintf("%s:%s", host, "443"), nil
			}
			return method, fmt.Sprintf("%s:%s", host, "80"), nil
		}
	}

	return method, "", errors.New("can not find host and port")
}

const (
	idVer           = 0
	idMethod        = 1
	socksVer5       = 5
	socksCmdConnect = 1
)

var (
	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support 1 method now")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")
)

func handleSS5(buf []byte, conn net.Conn) (hostPort string, err error) {
	if len(buf) > 258 {
		return "", errVer
	}
	if buf[idVer] != socksVer5 {
		return "", errVer
	}
	_, err = conn.Write([]byte{socksVer5, 0})

	hostPort, err = getRequest(conn)
	if err != nil {
		return "", err
	}
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	return
}

func getRequest(conn net.Conn) (host string, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4 = 1 // type is ipv4 address
		typeDm   = 3 // type is domain address
		typeIPv6 = 4 // type is ipv6 address

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)
	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)
	var n int
	// read till we get possible domain length field
	if n, err = io.ReadAtLeast(conn, buf, idDmLen+1); err != nil {
		return
	}
	// check version and cmd
	if buf[idVer] != socksVer5 {
		err = errVer
		return
	}
	if buf[idCmd] != socksCmdConnect {
		err = errCmd
		return
	}

	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
	default:
		err = errAddrType
		return
	}

	if n == reqLen {
		// common case, do nothing
	} else if n < reqLen { // rare case
		if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return
		}
	} else {
		err = errReqExtraData
		return
	}
	switch buf[idType] {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}
	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return
}
