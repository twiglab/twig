package twig

import (
	"context"
	"net/http"
)

// Server Http处理器接口
type Server interface {
	Attacher
	Cycler
}

type Work struct {
	*http.Server
	twig *Twig
}

func NewWork() *Work {
	return &Work{
		Server: &http.Server{
			Addr: DefaultAddress,
		},
	}
}

func (w *Work) Attach(t *Twig) {
	w.twig = t
	w.Handler = t
}

func (w *Work) Shutdown(ctx context.Context) error {
	return w.Server.Shutdown(ctx)
}

func (w *Work) Start() (err error) {
	go func() {
		if err = w.Server.ListenAndServe(); err != nil {
			w.twig.Logger.Panic(err)
		}
	}()
	return
}
