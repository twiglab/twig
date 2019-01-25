package twig

import (
	"context"
	"net/http"
)

type Server interface {
	Handler(http.Handler)
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

func (w *Work) Handler(h http.Handler) {
	w.Server.Handler = h
}

func (w *Work) Attach(twig *Twig) {
	w.twig = twig
}

func (w *Work) Shutdown(ctx context.Context) error {
	return w.Server.Shutdown(ctx)
}

func (w *Work) Start() (err error) {
	go func() {
		err = w.Server.ListenAndServe()
	}()
	return
}
