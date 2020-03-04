package server

import (
	"net"
	"os"
)

type debugConn struct {
	isDebug bool
	net.Conn
}

func (d debugConn) Write(p []byte) (int, error) {
	if d.isDebug {
		os.Stdout.Write(p)
	}
	return d.Conn.Write(p)
}

func (d debugConn) Read(p []byte) (int, error) {
	n, err := d.Conn.Read(p)
	if d.isDebug {
		os.Stdout.Write(p[:n])
	}
	return n, err
}
