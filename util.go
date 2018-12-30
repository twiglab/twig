package twig

import (
	"net/http"
	"net/http/pprof"
	"reflect"
	"runtime"
	"strings"
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

type Route struct {
	Name   string
	Path   string
	Method string
}

const XMLHttpRequest = "XMLHttpRequest"

func IsAJAX(r *http.Request) bool {
	return strings.Contains(r.Header.Get(HeaderXRequestedWith), XMLHttpRequest)
}
