package twig

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// 获取当前请求路径
func GetReqPath(r *http.Request) string {
	path := r.URL.RawPath

	if path == "" {
		path = r.URL.Path
	}

	return path
}

// 获取handler的名称
func HandlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// HelloTwig! ~~
func HelloTwig(c Ctx) error {
	return c.Stringf(http.StatusOK, "Hello %s!", "Twig")
}

// 包装handler
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

type RouteDesc struct {
	N string
	P string
	M string
}

func (r *RouteDesc) ID() string {
	return r.M + r.P
}

func (r *RouteDesc) Name() string {
	return r.N
}

func (r *RouteDesc) Method() string {
	return r.M
}

func (r *RouteDesc) Path() string {
	return r.P
}

func (r *RouteDesc) SetName(name string) {
	r.N = name
}

// 判断当前请求是否为AJAX
func IsAJAX(r *http.Request) bool {
	return strings.Contains(r.Header.Get(HeaderXRequestedWith), XMLHttpRequest)
}

// 设置关联关系
func attach(i interface{}, t *Twig) {
	if linker, ok := i.(Attacher); ok {
		linker.Attach(t)
	}
}

func GetPartner(id string, c Ctx) Partner {
	t := c.Twig()
	if p, ok := t.Partner(id); ok {
		return p
	}

	c.Logger().Panicf("Twig: Partner (%s) is not exist!", id)

	return nil
}

type Mounter interface {
	Mount(Register)
}

type Config struct {
	R Register
	N Nameder
}

func Cfg() *Config {
	return &Config{}
}

func (c *Config) WithRegister(r Register) *Config {
	c.R = r
	return c
}

func (c *Config) WithNameder(n Nameder) *Config {
	c.N = n
	return c
}

func (c *Config) With(r Register, n Nameder) *Config {
	c.R = r
	c.N = n
	return c
}

func (c *Config) SetName(name string) *Config {
	c.N.SetName(name)
	return c
}

func (c *Config) Use(m ...MiddlewareFunc) *Config {
	c.R.Use(m...)
	return c
}

func (c *Config) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Config {
	c.N = c.R.AddHandler(method, path, handler, m...)
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

func (c *Config) Done() {
	c.R = nil
	c.N = nil
	c = nil
}
