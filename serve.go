package twig

import (
	"context"
	"log"
	"net/http"
	"os"
)

const (
	DefaultAddress = ":4321"
)

type Starter interface {
	Start() error
}
type Shutdowner interface {
	Shutdown(context.Context) error
}

type Server interface {
	Starter
	Shutdowner
}

type Servant interface {
	Attach(http.Handler)
	Server
}

type ClassicServer struct {
	*http.Server
}

func NewClassicServer(addr string) *ClassicServer {
	address := addr
	if addr == "" {
		address = DefaultAddress
	}
	return &ClassicServer{
		Server: &http.Server{
			Addr:           address,
			ErrorLog:       log.New(os.Stderr, "twig-server-log-", log.LstdFlags|log.Llongfile),
			MaxHeaderBytes: defaultHeaderBytes,
		},
	}
}

func (s *ClassicServer) Attach(h http.Handler) {
	s.Handler = h
}

func (s *ClassicServer) Address(addr string) *ClassicServer {
	s.Addr = addr
	return s
}

func (s *ClassicServer) Start() error {
	go func() {
		if err := s.ListenAndServe(); err != nil {
			s.Server.ErrorLog.Println(err)
		}
	}()
	return nil
}

func (s *ClassicServer) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
