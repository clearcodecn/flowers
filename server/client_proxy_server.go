package server

import (
	"context"
	"fmt"
	"github.com/clearcodecn/flowers/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
	"sync"
)

type ClientProxyServer struct {
	opt *Option

	clientPool sync.Pool

	ln net.Listener
}

type ClientProxyServerOptions struct {
	Address       string
	ServerAddress string

	Password string
}

func NewClientProxyServer(opts ...Options) (*ClientProxyServer, error) {
	o := new(Option)
	for _, opt := range opts {
		opt(o)
	}
	s := new(ClientProxyServer)
	s.opt = o

	s.clientPool.New = func() interface{} {
		// TODO:: add tls support
		conn, _ := grpc.Dial(o.ServerProxyAddress, grpc.WithInsecure())
		return proto.NewProxyServiceClient(conn)
	}

	return s, nil
}

func (c *ClientProxyServer) GetClient() proto.ProxyServiceClient {
	return c.clientPool.Get().(proto.ProxyServiceClient)
}

func (c *ClientProxyServer) ReleaseClient(client proto.ProxyServiceClient) {
	c.clientPool.Put(client)
}

func (c *ClientProxyServer) Stop() error {
	return c.ln.Close()
}

func (c *ClientProxyServer) Run() error {
	ln, err := net.Listen("tcp", c.opt.ClientProxyAddress)
	if err != nil {
		return err
	}
	c.ln = ln

	logrus.Infof("client proxy running at: %s", c.opt.ClientProxyAddress)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			return err
		}
		go c.handleConn(conn)
	}
}

func (c *ClientProxyServer) handleConn(conn net.Conn) {
	var b = make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		logrus.Errorf("read local connection failed: %s", err)
		return
	}
	var (
		method, host string
	)
	host, err = c.handleSS5(b[:n], conn)
	if err == nil {
		b = b[:]
		n, err = conn.Read(b)
		if err != nil {
			return
		}
	} else {
		method, host, err = FindHost(b[:n])
		if err != nil {
			logrus.Errorf("parse hostPort failed: %s %s", err)
			return
		}
		if method == http.MethodConnect {
			_, _ = fmt.Fprint(conn, "HTTP/1.1 200 Connection established\r\n\r\n")
		}
	}
	client := c.GetClient()

	stream, err := client.Proxy(context.Background())
	if err != nil {
		logrus.Errorf("connect server failed: %s", err)
		return
	}

	logrus.Infof("proxy for: %s", host)

	req := &proto.Request{
		Host: host,
	}
	if err := stream.Send(req); err != nil {
		logrus.Infof("can not connect to remote server: %s", err)
		return
	}
	var (
		reqCh  = make(chan []byte, 10)
		respCh = make(chan []byte, 10)
		closed atomicBool
		o      sync.Once
	)
	if http.MethodConnect != method {
		reqCh <- b[:n]
	}

	var closeFunc = func() {
		o.Do(func() {
			logrus.Infof("closed %s", host)
			closed.SetTrue()
			close(reqCh)
			close(respCh)
			stream.CloseSend()
		})

		if err := recover(); err != nil {
			logrus.Errorf("panic: %s", err)
		}
	}
	go func() {
		defer closeFunc()

		for b := range reqCh {
			if closed.Bool() {
				return
			}
			if err := stream.Send(&proto.Request{
				Body: b,
			}); err != nil {
				logrus.Errorf("can not request to server: %s", err)
				return
			}
		}
	}()
	go func() {
		defer closeFunc()

		for b := range respCh {
			if closed.Bool() {
				return
			}
			if _, err := conn.Write(b); err != nil {
				return
			}
		}
	}()
	go func() {
		defer closeFunc()
		for {
			if closed.Bool() {
				return
			}
			var b = make([]byte, 1024)
			n, err := conn.Read(b)
			if err != nil {
				return
			}
			if closed.Bool() {
				return
			}
			reqCh <- b[:n]
		}
	}()
	go func() {
		defer closeFunc()
		for {
			if closed.Bool() {
				return
			}
			resp, err := stream.Recv()
			if err != nil {
				return
			}
			if closed.Bool() {
				return
			}
			respCh <- resp.Body
		}
	}()
}
