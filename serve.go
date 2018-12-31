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

type Server interface {
	Start() error
	Shutdown(context.Context) error

	Attacher
}

type Servant struct {
	Server *http.Server
	t      *Twig
}

func NewServnat(addr string) *Servant {
	address := addr
	if addr == "" {
		address = DefaultAddress
	}
	return &Servant{
		Server: &http.Server{
			Addr:           address,
			ErrorLog:       log.New(os.Stderr, "twig-server-log-", log.LstdFlags|log.Llongfile),
			MaxHeaderBytes: defaultHeaderBytes,
		},
	}
}

func (s *Servant) Attach(t *Twig) {
	s.Server.Handler = t
	s.t = t
}
func (s *Servant) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

func (s *Servant) Start() error {
	return s.Server.ListenAndServe()
}

func HttpServerWrap(s *http.Server) Server {
	return &Servant{
		Server: s,
	}
}
