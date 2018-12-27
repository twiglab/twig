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

type Servant interface {
	Start() error
	Shutdown(context.Context) error

	Attacher
}

type ClassicServer struct {
	*http.Server
	t       *Twig
	address string
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
		address: addr,
	}
}

func (s *ClassicServer) Attach(t *Twig) {
	s.Handler = t
	s.t = t
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
