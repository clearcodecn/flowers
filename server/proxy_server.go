package server

import (
	"context"
	"flowers/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ProxyServer struct {
	opt *Option
	wg  sync.WaitGroup

	gs *grpc.Server
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
		s.wg.Wait()
		cancel()
	}()

	<-ctx.Done()
	return nil
}

func (s *ProxyServer) Proxy(stream proto.ProxyService_ProxyServer) error {
	s.wg.Add(1)
	defer s.wg.Done()

	req, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "recv stream failed")
	}
	var (
		conn net.Conn
	)
	if s.opt.Cipher != nil {
		b, err := s.opt.Cipher.Decode([]byte(req.Host))
		if err != nil {
			logrus.Errorf("cipher data failed: %s", err)
			return errors.New("invalid password")
		}
		req.Host = string(b)
	}

	if req.Host == "" {
		return errors.New("invalid host")
	}
	if req.Host != "" {
		// fist package
		conn, err = net.Dial("tcp", req.Host)
		if err != nil {
			return err
		}
		logrus.Infof("connected to: %s", req.Host)
	} else {
		return errors.Errorf("invalid host: %s", req.Host)
	}
	return s.pipe(stream, conn)
}

func (s *ProxyServer) pipe(stream proto.ProxyService_ProxyServer, conn net.Conn) error {
	var (
		reqChan        = make(chan []byte, 10)
		respChan       = make(chan []byte, 10)
		closed   int64 = 0
	)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for b := range reqChan {
			if s.opt.Cipher != nil {
				var err error
				b, err = s.opt.Cipher.Decode(b)
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
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for b := range respChan {
			if s.opt.Cipher != nil {
				var err error
				b, err = s.opt.Cipher.Encode(b)
				if err != nil {
					logrus.Errorf("cipher data failed: %s", err)
					return
				}
			}
			if err := stream.Send(&proto.Response{
				Body: b,
			}); err != nil {
				return
			}
		}
	}()
	// recv
	s.wg.Add(1)
	go func() {
		defer func() {
			close(reqChan)
			close(respChan)
			atomic.StoreInt64(&closed, 1)
			conn.Close()
			s.wg.Done()
		}()
		for {
			req, err := stream.Recv()
			if err != nil {
				return
			}
			reqChan <- req.Body
		}
	}()
	for {
		var b = make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			return err
		}
		if atomic.LoadInt64(&closed) == 0 {
			respChan <- b[:n]
		}
	}
}
