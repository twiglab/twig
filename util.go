package twig

import (
	"net/http"
	"reflect"
	"runtime"
)

func GetReqPath(r *http.Request) string {
	path := r.URL.RawPath

	if path == "" {
		path = r.URL.Path
	}

	return path
}

func HandlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

type ToyMux struct{}

func NewToyMux() *ToyMux {
	return new(ToyMux)
}

func (m *ToyMux) Lookup(r *http.Request, c MContext) {
	c.SetHandler(func(c C) error {
		return c.String(http.StatusOK, "Hello Twig!")
	})
}

type Group struct {
}

func NewGroup() *Group {
	return nil
}

type MutiMux map[string]Muxer

func (m MutiMux) Lookup(r *http.Request, mc MContext) {
	s := r.Header.Get("")
	mux := m[s]
	mux.Lookup(r, mc)
}
