package twig

import (
	"net/http"
)

// M 全局通用的map
type M map[string]interface{}

// Mouter 接口用于模块化设置路由
type Mounter interface {
	Mount(Register)
}

// 获取当前请求路径
func GetReqPath(r *http.Request) string {
	path := r.URL.RawPath

	if path == "" {
		path = r.URL.Path
	}

	return path
}

// Attach 设置关联关系
func Attach(i interface{}, t *Twig) {
	if attacher, ok := i.(Attacher); ok {
		attacher.Attach(t)
	}
}

type Cfg struct {
	R Register
	N Namer
}

func Config(r Register) *Cfg {
	return &Cfg{
		R: r,
		N: nil,
	}
}

func (c *Cfg) WithNamer(n Namer) *Cfg {
	c.N = n
	return c
}

func (c *Cfg) SetName(name string) *Cfg {
	c.N.SetName(name)
	return c
}

func (c *Cfg) Use(m ...MiddlewareFunc) *Cfg {
	c.R.Use(m...)
	return c
}

func (c *Cfg) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	c.N = c.R.AddHandler(method, path, handler, m...)
	return c
}

func (c *Cfg) Get(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(GET, path, handler, m...)
}

func (c *Cfg) Post(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(POST, path, handler, m...)
}

func (c *Cfg) Delete(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(DELETE, path, handler, m...)
}

func (c *Cfg) Put(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(PUT, path, handler, m...)
}

func (c *Cfg) Patch(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(PATCH, path, handler, m...)
}

func (c *Cfg) Head(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(HEAD, path, handler, m...)
}

func (c *Cfg) Options(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(OPTIONS, path, handler, m...)
}

func (c *Cfg) Trace(path string, handler HandlerFunc, m ...MiddlewareFunc) *Cfg {
	return c.AddHandler(TRACE, path, handler, m...)
}

func (c *Cfg) Mount(mount Mounter) *Cfg {
	mount.Mount(c.R)
	c.N = nil
	return c
}

func (c *Cfg) Static(path, file string, m ...MiddlewareFunc) *Cfg {
	return c.Get(path, Static(file), m...)
}

func (c *Cfg) Done() {
	c.R = nil
	c.N = nil
}

// Group 提供理由分组支持
type Group struct {
	prefix string
	m      []MiddlewareFunc
	reg    Register
}

func NewGroup(r Register, prefix string) *Group {
	return &Group{
		prefix: prefix,
		reg:    r,
	}
}

func (g *Group) Use(mid ...MiddlewareFunc) {
	g.m = append(g.m, mid...)
}

func (g *Group) AddHandler(method, path string, h HandlerFunc, m ...MiddlewareFunc) Route {
	name := HandlerName(h)
	handler := Merge(h, g.m)

	route := g.reg.AddHandler(method, g.prefix+path, handler, m...)

	route.SetName(name)

	return route
}
