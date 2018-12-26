package twig

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type HttpErrorHandler func(error, C)
type H map[string]interface{}

type Identifier interface {
	Name() string
	Type() string
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
		Muxer: NewToyMux(),
	}
	t.Logger = newLog(os.Stdout, "twig-log-")
	t.HttpErrorHandler = t.DefaultHttpErrorHandler
	t.pool.New = func() interface{} {
		return t.newCtx(nil, nil)
	}

	return t
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

func (t *Twig) DefaultHttpErrorHandler(err error, c C) {

	var code int = http.StatusInternalServerError
	var msg interface{}

	if e, ok := err.(*HttpError); ok {
		code = e.Code
		msg = e.Msg

		if e.Internal != nil {
			err = fmt.Errorf("%v, %v", err, e.Internal)
		}
	} else {
		msg = http.StatusText(code)
	}

	if m, ok := msg.(string); ok {
		msg = map[string]string{"msg": m}
	}

	if !c.Resp().Committed {
		if c.Req().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.Json(code, msg)
		}
		if err != nil {
			t.Logger.Println(err)
		}
	}
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
