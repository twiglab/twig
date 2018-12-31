package twig

import (
	"context"
	"log"
	"net/http"
	"os"
)

type Server interface {
	Start() error
	Shutdown(context.Context) error

	Assocer
}

type Servant struct {
	Server *http.Server
	t      *Twig
}

func DefaultServnat() *Servant {
	return &Servant{
		Server: &http.Server{
			Addr:           DefaultAddress,
			ErrorLog:       log.New(os.Stderr, "twig-server-log-", log.LstdFlags|log.Llongfile),
			MaxHeaderBytes: defaultHeaderBytes,
		},
	}
}

func (s *Servant) Assoc(t *Twig) {
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
