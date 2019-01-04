package twig

import (
	"context"
	"log"
	"net/http"
	"os"
)

// Server 接口
type Server interface {
	Cycler
	Attacher
}

// Servant 默认的Sever实现用于Http处理
type Servant struct {
	Server *http.Server
	t      *Twig
}

func DefaultServant() *Servant {
	return &Servant{
		Server: &http.Server{
			Addr:           DefaultAddress,
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

func (s *Servant) Start() (err error) {
	go func() {
		err = s.Server.ListenAndServe()
	}()
	return
}

func HttpServerWrap(s *http.Server) Server {
	return &Servant{
		Server: s,
	}
}
