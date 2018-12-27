package twig

import (
	"context"
	"net/http"
	"os"
	"sync"
)

type H map[string]interface{}

type Attacher interface {
	Attach(*Twig)
}

type Identifier interface {
	Name() string
	Type() string
	Desc() string
}

type Twig struct {
	HttpErrorHandler HttpErrorHandler

	Logger  Logger
	Muxer   Muxer
	Servant Servant

	Debug bool

	pre []MiddlewareFunc
	mid []MiddlewareFunc

	pool sync.Pool
}

func TODO() *Twig {
	t := &Twig{
		Debug: false,
	}
	t.pool.New = func() interface{} {
		return t.newCtx(nil, nil)
	}
	t.WithServant(NewClassicServer(DefaultAddress)).
		WithHttpErrorHandler(DefaultHttpErrorHandler).
		WithLogger(newLog(os.Stdout, "twig-log-")).
		WithMux(NewToyMux())
	return t
}

func (t *Twig) WithLogger(l Logger) *Twig {
	t.Logger = l
	return t
}

func (t *Twig) WithHttpErrorHandler(eh HttpErrorHandler) *Twig {
	t.HttpErrorHandler = eh
	return t
}

func (t *Twig) Pre(m ...MiddlewareFunc) *Twig {
	t.pre = append(t.pre, m...)
	return t
}

func (t *Twig) Use(m ...MiddlewareFunc) *Twig {
	t.mid = append(t.mid, m...)
	return t
}

func (t *Twig) WithMux(m Muxer) *Twig {
	t.Muxer = m
	m.Attach(t)
	return t
}

func (t *Twig) WithServant(s Servant) *Twig {
	t.Servant = s
	s.Attach(t)
	return t
}

func (t *Twig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := t.pool.Get().(*ctx)
	c.Reset(w, r)

	h := Enhance(func(ctx C) error {
		t.Muxer.Lookup(r, c)
		handler := Enhance(c.Handler(), t.mid)
		return handler(c)
	}, t.pre)

	if err := h(c); err != nil {
		t.HttpErrorHandler(err, c)
	}

	t.pool.Put(c)
}

func (t *Twig) Start() error {
	t.Logger.Println(banner)
	return t.Servant.Start()
}

func (t *Twig) Shutdown(ctx context.Context) error {
	return t.Servant.Shutdown(ctx)
}

func (t *Twig) newCtx(w http.ResponseWriter, r *http.Request) C {
	return &ctx{
		req:     r,
		resp:    NewResponseWarp(w),
		t:       t,
		store:   make(H),
		pvalues: make([]string, MaxParam),
		handler: NotFoundHandler,
	}
}
