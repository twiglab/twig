package twig

import (
	"context"
	"net/http"
	"strings"
)

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

// 判断当前请求是否为AJAX
func IsAJAX(r *http.Request) bool {
	return strings.Contains(r.Header.Get(HeaderXRequestedWith), XMLHttpRequest)
}

// Reverse 根据path和参数构建url
/*
func Reverse(path string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' && n < ln {
			for ; i < l && path[i] != '/'; i++ {
			}
			uri.WriteString(fmt.Sprintf("%v", params[n]))
			n++
		}
		if i < l {
			uri.WriteByte(path[i])
		}
	}
	return uri.String()
}
*/

// Attach 设置关联关系
func Attach(i interface{}, t *Twig) {
	if attacher, ok := i.(Attacher); ok {
		attacher.Attach(t)
	}
}

func Start(i interface{}) error {
	if cycler, ok := i.(Cycler); ok {
		return cycler.Start()
	}

	return nil
}

func Shutdown(i interface{}, c context.Context) error {
	if cycler, ok := i.(Cycler); ok {
		return cycler.Shutdown(c)
	}

	return nil
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

func (c *Cfg) Done() {
	c.R = nil
	c.N = nil
	c = nil
}
