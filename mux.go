package twig

import (
	"net/http"
)

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

// Muxer Muxer 描述一个具体的路由
type Muxer interface {
	Lookuper
	Register
}

// Matcher Matcher接口用于路由匹配
type Matcher interface {
	// Match 根据当前请求返回Lookuper， 如果不匹配，返回nil
	Match(*http.Request) Lookuper
}

type Muxes struct {
	Macthers []Matcher
	Default  Lookuper
}

func MutiMuxes(def Lookuper, matchers ...Matcher) *Muxes {
	return &Muxes{
		Macthers: matchers,
		Default:  def,
	}
}

func TwoMuxes(def Lookuper, matcher Matcher) *Muxes {
	return MutiMuxes(def, matcher)
}

func (w *Muxes) Lookup(method, path string, r *http.Request) Ctx {
	var lookuper Lookuper
	for _, m := range w.Macthers {
		if lookuper = m.Match(r); lookuper != nil {
			return lookuper.Lookup(method, path, r)
		}
	}
	return w.Default.Lookup(method, path, r)
}
func (w *Muxes) AddHandler(string, string, HandlerFunc, ...MiddlewareFunc) Router {
	panic("twig:MutiMux not supports AddHandler function")
}

func (w *Muxes) Use(...MiddlewareFunc) {
	panic("twig:MutiMux not supports Use function")
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

// ID 命名路由的ID
func (r *NamedRoute) ID() string {
	return r.M + r.P
}

// Name 命名路由的名称
func (r *NamedRoute) Name() string {
	return r.N
}

// Method 命名路由的Http方法
func (r *NamedRoute) Method() string {
	return r.M
}

// Path 命名路由的注册路径
func (r *NamedRoute) Path() string {
	return r.P
}

// SetName  Namer#SetName
func (r *NamedRoute) SetName(name string) {
	r.N = name
}

// Type 类型
func (r *NamedRoute) Type() string {
	return "handler"
}
