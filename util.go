package twig

import (
	"net/http"
)

type ExRegister interface {
	Register
	UsePlugger(...Plugger)
	GetPlugger(string) Plugger
}

// Mouter 接口用于模块化设置路由
type Mounter interface {
	Mount(ExRegister)
}

type MountFunc func(ExRegister)

func (m MountFunc) Mount(r ExRegister) {
	m(r)
}

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

type target struct {
	Register
	*Twig
}

// Use 消除冲突
func (t *target) Use(m ...MiddlewareFunc) {
	t.Register.Use(m...)
}

func NewExRegister(r Register, twig *Twig) ExRegister {
	return &target{
		Register: r,
		Twig:     twig,
	}
}

// Config Twig路由配置工具
type Config struct {
	r ExRegister
	n Namer
}

func NewConfig(r Register) *Config {
	return TwigConfig(r, nil)
}

func TwigConfig(r Register, twig *Twig) *Config {
	return NewConfigEx(NewExRegister(r, twig))
}

func NewConfigEx(r ExRegister) *Config {
	return &Config{
		r: r,
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

// Mount 挂载Mounter到当前ExRegister
func (c *Config) Mount(mount Mounter) *Config {
	mount.Mount(c.r)
	c.n = nil
	return c
}

// Static 增加静态路由
func (c *Config) Static(path, file string, m ...MiddlewareFunc) *Config {
	return c.Get(path, Static(file), m...)
}

// Group 配置路由组
func (c *Config) Group(path string, f MountFunc) *Config {
	f(NewGroup(c.r, path))
	return c
}

/*
	web.Config().
		Group("/api", func(r twig.ExRegister) {
			twig.NewConfigEx(r).
				Post("/addUser", func(c twig.Ctx) error {
					...
				})
		})
*/
// Group 提供理由分组支持
type Group struct {
	prefix string
	m      []MiddlewareFunc
	ExRegister
}

func NewGroup(r ExRegister, prefix string) *Group {
	return &Group{
		prefix:     prefix,
		ExRegister: r,
	}
}

func (g *Group) Use(mid ...MiddlewareFunc) {
	g.m = append(g.m, mid...)
}

func (g *Group) AddHandler(method, path string, h HandlerFunc, m ...MiddlewareFunc) Router {
	name := HandlerName(h)
	handler := Merge(h, g.m)

	route := g.ExRegister.AddHandler(method, g.prefix+path, handler, m...)

	route.SetName(name)

	return route
}
