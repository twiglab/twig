package twig

import (
	"net/http"
)

type Register interface {
	AddHandler(string, string, HandlerFunc, ...MiddlewareFunc) Route
	Use(...MiddlewareFunc)
}

type Lookuper interface {
	Lookup(string, string, *http.Request, Ctx)
}

//Muxer 接口
type Muxer interface {
	Lookuper
	Register
}

// Route 接口
type Route interface {
	Name() string
	ID() string
	Method() string
	Path() string
	Nameder
}
