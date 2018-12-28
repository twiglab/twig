package twig

import (
	"net/http"
	"net/http/pprof"
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

func HelloTwig(c Ctx) error {
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

func (m *ToyMux) Lookup(method, path string, r *http.Request, c Ctx) {
	c.SetHandler(HelloTwig)
}

func Enhance(handler HandlerFunc, m []MiddlewareFunc) HandlerFunc {
	if m == nil {
		return handler
	}

	h := handler
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h

}

var PprofIndex = WrapHttpHandlerFunc(pprof.Index)
var PprofCmdLine = WrapHttpHandlerFunc(pprof.Cmdline)
var PprofProfile = WrapHttpHandlerFunc(pprof.Profile)
var PprofSymbol = WrapHttpHandlerFunc(pprof.Symbol)
var PprofTrace = WrapHttpHandlerFunc(pprof.Trace)
