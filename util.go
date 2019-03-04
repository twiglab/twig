package twig

import (
	"net/http"
)

type PluginProvider interface {
	UsePlugger(...Plugger)
	GetPlugger(string) Plugger
}

type ExRegister interface {
	Register
	PluginProvider
}

// Mouter 接口用于模块化设置路由
type Mounter interface {
	Mount(ExRegister)
}

/*
type MountFunc func(ExRegister)

func (m MountFunc) Mount(r ExRegister) {
	m(r)
}
*/

// M 全局通用的map
type M map[string]interface{}

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

// Config Twig路由配置工具
type Config struct {
	r Register
	p PluginProvider
	n Namer
}

func newConfig(r Register, t *Twig) *Config {
	return &Config{
		r: r,
		p: t,
		n: t,
	}
}

// SetName 设置当前Namer的名称
func (c *Config) SetName(name string) *Config {
	c.n.SetName(name)
	return c
}

// Use 当前Register增加中间件
func (c *Config) Use(m ...MiddlewareFunc) *Config {
	c.Use(m...)
	return c
}

// AddHandler 增加Handler
func (c *Config) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	c.n = c.r.AddHandler(method, path, handler, m...)
	return c
}

func (c *Config) Get(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(GET, path, handler, m...)
}

func (c *Config) Post(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(POST, path, handler, m...)
}

func (c *Config) Delete(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(DELETE, path, handler, m...)
}

func (c *Config) Put(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(PUT, path, handler, m...)
}

func (c *Config) Patch(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(PATCH, path, handler, m...)
}

func (c *Config) Head(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(HEAD, path, handler, m...)
}

func (c *Config) Options(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(OPTIONS, path, handler, m...)
}

func (c *Config) Trace(path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	return c.AddHandler(TRACE, path, handler, m...)
}

// Mount 挂载Mounter到当前Register
func (c *Config) Mount(mount Mounter) *Config {
	mount.Mount(
		&struct {
			Register
			PluginProvider
		}{
			c.r,
			c.p,
		},
	)
	c.n = nil
	return c
}

// Static 增加静态路由
func (c *Config) Static(path, file string, m ...MiddlewareFunc) *Config {
	return c.Get(path, Static(file), m...)
}

// Group 配置路由组
// 存在缺陷，Group不支持Plugin(TODO)
func (c *Config) Group(path string, f func(Register)) *Config {
	f(NewGroup(c.r, path))
	return c
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

func (g *Group) AddHandler(method, path string, h HandlerFunc, m ...MiddlewareFunc) Router {
	name := HandlerName(h)
	handler := Merge(h, g.m)

	route := g.reg.AddHandler(method, g.prefix+path, handler, m...)

	route.SetName(name)

	return route
}
