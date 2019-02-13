package twig

import (
	"net/http"
)

// Matcher Matcher接口用于路由匹配
type Matcher interface {
	Match(*http.Request, *Twig) Lookuper
}
type MatcherFunc func(*http.Request, *Twig) Lookuper

func (m MatcherFunc) Match(r *http.Request, t *Twig) Lookuper {
	return m(r, t)
}

// Register 接口
// 实现路由注册
type Register interface {
	AddHandler(string, string, HandlerFunc, ...MiddlewareFunc) Router
	Use(...MiddlewareFunc)
}

// Lookuper 接口
// 实现路由查找
type Lookuper interface {
	Lookup(string, string, *http.Request) Ctx
}

type Muxer interface {
	Lookuper
	Register
}

type Wrapper interface {
	Matcher
	Configer
}

// Router 接口，Route接口用于描述一个已经加入Register的路由，由Register的AddHandler方法返回
// Router 提供命名路由的方法，被命名的路由可以用于Ctx.URL方法查找
type Router interface {
	Identifier
	Method() string
	Path() string
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

func (r *NamedRoute) Type() string {
	return "handler"
}
