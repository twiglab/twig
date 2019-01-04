package twig

import (
	"net/http"
)

// Register 接口
type Register interface {
	AddHandler(string, string, HandlerFunc, ...MiddlewareFunc) Route
	Use(...MiddlewareFunc)
}

// Lookuper 接口
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
	Identifier
	Method() string
	Path() string
	Name() string
	Namer
}

type NamedRoute struct {
	N string // name
	P string // path
	M string // method
}

func (r *NamedRoute) ID() string {
	return r.M + r.P
}

func (r *NamedRoute) Name() string {
	return r.N
}

func (r *NamedRoute) Method() string {
	return r.M
}

func (r *NamedRoute) Path() string {
	return r.P
}

func (r *NamedRoute) SetName(name string) {
	r.N = name
}
