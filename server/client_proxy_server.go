package server

import (
	"context"
	"flowers/proto"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ClientProxyServer struct {
	opt *Option

	client proto.ProxyServiceClient

	ln net.Listener

	wg sync.WaitGroup
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

	// TODO:: add tls support
	conn, err := grpc.Dial(o.ServerProxyAddress, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "can not create grpc connection")
	}

	s.client = proto.NewProxyServiceClient(conn)

	return s, nil
}
func (c *ClientProxyServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		c.wg.Wait()
		cancel()
	}()
	<-ctx.Done()
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
			logrus.Errorf("parse hostPort failed: %s", err)
			return
		}
		if method == http.MethodConnect {
			_, _ = fmt.Fprint(conn, "HTTP/1.1 200 Connection established\r\n\r\n")
		}
	}
	client, err := c.client.Proxy(context.Background())
	if err != nil {
		logrus.Errorf("connect server failed: %s", err)
		return
	}

	logrus.Infof("proxy for: %s", host)

	if c.opt.Cipher != nil {
		b, err := c.opt.Cipher.Encode([]byte(host))
		if err != nil {
			logrus.Errorf("cipher data failed: %s", err)
			return
		}
		host = string(b)
	}

	req := &proto.Request{
		Host: host,
	}
	if err := client.Send(req); err != nil {
		logrus.Infof("can not connect to remote server: %s", err)
		return
	}
	c.wg.Add(4)
	var (
		reqCh        = make(chan []byte, 10)
		respCh       = make(chan []byte, 10)
		closed int64 = 0
	)
	if http.MethodConnect != method {
		reqCh <- b[:n]
	}

	go func() {
		defer c.wg.Done()
		for b := range reqCh {
			if c.opt.Cipher != nil {
				var err error
				b, err = c.opt.Cipher.Encode(b)
				if err != nil {
					logrus.Errorf("cipher data failed: %s", err)
					return
				}
			}
			if err := client.Send(&proto.Request{
				Body: b,
			}); err != nil {
				logrus.Errorf("can not request to server: %s", err)
				return
			}
		}
	}()
	go func() {
		defer c.wg.Done()
		for b := range respCh {
			if c.opt.Cipher != nil {
				var err error
				b, err = c.opt.Cipher.Decode(b)
				if err != nil {
					logrus.Errorf("cipher data failed: %s", err)
					return
				}
			}
			if _, err := conn.Write(b); err != nil {
				return
			}
		}
	}()
	go func() {
		defer func() {
			c.wg.Done()
			close(reqCh)
			close(respCh)
			atomic.StoreInt64(&closed, 1)
			conn.Close()
			client.CloseSend()
		}()
		for {
			var b = make([]byte, 1024)
			n, err := conn.Read(b)
			if err != nil {
				return
			}
			reqCh <- b[:n]
		}
	}()
	go func() {
		defer c.wg.Done()
		for {
			resp, err := client.Recv()
			if err != nil {
				return
			}
			if atomic.LoadInt64(&closed) == 0 {
				respCh <- resp.Body
			}
		}
	}()
}
