package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ClientHttpServer struct {
	engine  *gin.Engine
	handler *http.Server

	opt *Option
}

func NewClientHttpServer(opts ...Options) {
	var opt = new(Option)
	for _, o := range opts {
		o(opt)
	}
	s := new(ClientHttpServer)
	s.opt = opt
	engine := gin.Default()

	s.engine = engine

	s.setupRoute()
}

func (s *ClientHttpServer) setupRoute() {
	// TODO::
}

func (s *ClientHttpServer) Run() error {
	srv := &http.Server{
		Addr:    s.opt.ClientHttpAddress,
		Handler: s.engine,
	}
	s.handler = srv
	return srv.ListenAndServe()
}

func (s *ClientHttpServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.handler.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
