package twig

import (
	"net/http"
)

// Register
type Register interface {
	AddHandler(string, string, HandlerFunc, ...MiddlewareFunc) Route
	Use(...MiddlewareFunc)
}

// Lookuper
type Lookuper interface {
	Lookup(string, string, *http.Request, MCtx)
}

// Muxer 接口
type Muxer interface {
	Lookuper
	Register
}

// Route 接口，Route接口用于描述一个已经加入Register的路由，由Register的AddHandler方法返回
// Route 提供命名路由的方法，被命名的路由可以用于Ctx.URL方法查找
type Route interface {
	Name() string
	ID() string
	Method() string
	Path() string
	Namer
}
