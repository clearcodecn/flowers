package server

import (
	"context"
	"github.com/clearcodecn/flowers/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type ProxyServer struct {
	opt *Option
	gs  *grpc.Server
}

func NewProxyServer(opts ...Options) *ProxyServer {
	var o = new(Option)
	for _, opt := range opts {
		opt(o)
	}

	s := new(ProxyServer)
	s.opt = o

	return s
}

func (s *ProxyServer) Run() error {
	gs := grpc.NewServer()
	proto.RegisterProxyServiceServer(gs, s)
	ln, err := net.Listen("tcp", s.opt.ServerProxyAddress)
	if err != nil {
		return err
	}
	s.gs = gs
	logrus.Infof("proxy server listen at: %s", s.opt.ServerProxyAddress)
	return gs.Serve(ln)
}

func (s *ProxyServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		s.gs.GracefulStop()
		cancel()
	}()

	<-ctx.Done()
	return nil
}

func (s *ProxyServer) Proxy(stream proto.ProxyService_ProxyServer) error {
	req, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "recv stream failed")
	}
	var (
		conn net.Conn
	)
	if req.Host == "" {
		return errors.New("invalid host")
	}
	if req.Host != "" {
		// fist package
		conn, err = net.Dial("tcp", req.Host)
		if err != nil {
			logrus.Errorf("can not dial: %s", req.Host)
			return err
		}
		logrus.Infof("connected to: %s", req.Host)
	} else {
		return errors.Errorf("invalid host: %s", req.Host)
	}
	done := make(chan struct{})
	go s.handleProxy(stream, conn, done)
	<-done

	logrus.Infof("close conn proxy for: ", req.Host)
	return nil
}

func (s *ProxyServer) handleProxy(stream proto.ProxyService_ProxyServer, conn net.Conn, done chan struct{}) {
	var (
		reqChan  = make(chan []byte, 10)
		respChan = make(chan []byte, 10)
		closed   atomicBool
		o        sync.Once
	)
	var closeFunc = func() {
		o.Do(func() {
			conn.Close()
			closed.SetTrue()
			close(reqChan)
			close(respChan)
			close(done)
		})
		if err := recover(); err != nil {
			logrus.Errorf("panic: %s", err)
		}
	}
	go func() {
		defer closeFunc()
		for b := range reqChan {
			if _, err := conn.Write(b); err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				logrus.Errorf("write conn err: %s", err)
			}
		}
	}()
	go func() {
		defer closeFunc()
		for b := range respChan {
			if err := stream.Send(&proto.Response{
				Body: b,
			}); err != nil {
				if err == io.EOF {
					return
				}
				logrus.Errorf("send err: %s", err)
				continue
			}
		}
	}()
	// recv
	go func() {
		defer closeFunc()
		for {
			if closed.Bool() {
				return
			}
			req, err := stream.Recv()
			if err != nil {
				logrus.Errorf("recv err: %s", err)
				continue
			}
			if closed.Bool() {
				return
			}
			reqChan <- req.Body
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
				if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				logrus.Errorf("can not read from remote: %s", err)
				continue
			}
			if closed.Bool() {
				return
			}
			respChan <- b[:n]
		}
	}()
}
