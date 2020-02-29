package server

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
)

//FindHost
func FindHost(data []byte) (method string, host string, err error) {
	arr := strings.Split(string(data), "\r\n")
	if len(arr) < 2 {
		err = errors.New("invalid http request part")
		return
	}
	{
		part := arr[0]
		partArr := strings.Split(part, " ")
		method = partArr[0]
	}
	{
		// Host:
		host = arr[1][6:]
		if !strings.Contains(host, ":") {
			if method == http.MethodConnect {
				host = net.JoinHostPort(host, "443")
			} else {
				host = net.JoinHostPort(host, "80")
			}
		}
	}
	return
}

type atomicBool int32

func (b *atomicBool) Bool() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) SetTrue()   { atomic.StoreInt32((*int32)(b), 1) }
