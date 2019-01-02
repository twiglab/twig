package twig

import (
	"bytes"
	"fmt"
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

//根据path和参数构建url
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

// HelloTwig! ~~
func HelloTwig(c Ctx) error {
	return c.Stringf(http.StatusOK, "Hello %s!", "Twig")
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

type Conf struct {
	R Register
	N Namer
}

func Config(r Register) *Conf {
	return &Conf{
		R: r,
		N: nil,
	}
}

func (c *Conf) WithNamer(n Namer) *Conf {
	c.N = n
	return c
}

func (c *Conf) SetName(name string) *Conf {
	c.N.SetName(name)
	return c
}

func (c *Conf) Use(m ...MiddlewareFunc) *Conf {
	c.R.Use(m...)
	return c
}

func (c *Conf) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	c.N = c.R.AddHandler(method, path, handler, m...)
	return c
}

func (c *Conf) Get(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(GET, path, handler, m...)
}

func (c *Conf) Post(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(POST, path, handler, m...)
}

func (c *Conf) Delete(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(DELETE, path, handler, m...)
}

func (c *Conf) Put(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(PUT, path, handler, m...)
}

func (c *Conf) Patch(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(PATCH, path, handler, m...)
}

func (c *Conf) Head(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(HEAD, path, handler, m...)
}

func (c *Conf) Options(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(OPTIONS, path, handler, m...)
}

func (c *Conf) Trace(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(TRACE, path, handler, m...)
}

func (c *Conf) Done() {
	c.R = nil
	c.N = nil
	c = nil
}
