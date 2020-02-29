package server

import (
	"errors"
	"net"
	"net/http"
	"strings"
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
