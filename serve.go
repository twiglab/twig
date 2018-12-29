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

type DefaultServer struct {
	*http.Server
	t       *Twig
	address string
}

func NewClassicServer(addr string) *DefaultServer {
	address := addr
	if addr == "" {
		address = DefaultAddress
	}
	return &DefaultServer{
		Server: &http.Server{
			Addr:           address,
			ErrorLog:       log.New(os.Stderr, "twig-server-log-", log.LstdFlags|log.Llongfile),
			MaxHeaderBytes: defaultHeaderBytes,
		},
		address: addr,
	}
}

func (s *DefaultServer) Attach(t *Twig) {
	s.Handler = t
	s.t = t
}

func (s *DefaultServer) Start() error {
	return s.ListenAndServe()
}

func (s *DefaultServer) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
