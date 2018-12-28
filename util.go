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

func HelloTwig(c C) error {
	return c.String(http.StatusOK, "Hello Twig!")
}

type ToyMux struct {
	t *Twig
}

func NewToyMux() *ToyMux {
	return new(ToyMux)
}

func (m *ToyMux) Attach(t *Twig) {
	m.t = t
}

func (m *ToyMux) Lookup(method, path string, r *http.Request, c C) {
	c.SetHandler(HelloTwig)
}

type Group struct {
}

func NewGroup() *Group {
	return nil
}
