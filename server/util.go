package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

//FindHost
func FindHost(data []byte) (method string, host string, err error) {

	fmt.Println(string(data))

	arr := strings.Split(string(data), "\r\n")
	part := arr[0]
	partArr := strings.Split(part, " ")
	method = partArr[0]
	if len(partArr) < 2 {
		return "", "", errors.New("can not parse host:port in first header")
	}
	hostPort := partArr[1]
	if strings.Contains(hostPort, ":") {
		return method, hostPort, nil
	}
	if method == http.MethodConnect {
		hostPort = hostPort + ":443"
	} else {
		hostPort = hostPort + ":80"
	}
	return method, hostPort, nil
}

type atomicBool int32

func (b *atomicBool) Bool() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) SetTrue()   { atomic.StoreInt32((*int32)(b), 1) }
