package twig

import (
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
)

type HandlerFunc func(Ctx) error

func (h HandlerFunc) Mount(method, path string, reg Register, m ...MiddlewareFunc) Route {
	return reg.AddHandler(method, path, h, m...)
}

type MiddlewareFunc func(HandlerFunc) HandlerFunc

func (m MiddlewareFunc) UsedBy(reg Register) {
	reg.Use(m)
}

func WrapHttpHandler(h http.Handler) HandlerFunc {
	return func(c Ctx) error {
		h.ServeHTTP(c.Resp(), c.Req())
		return nil
	}
}

func WrapHttpHandlerFunc(h http.HandlerFunc) HandlerFunc {
	return func(c Ctx) error {
		h.ServeHTTP(c.Resp(), c.Req())
		return nil
	}
}

func WrapMiddleware(m func(http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Ctx) (err error) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetReq(r)
				err = next(c)
			})).ServeHTTP(c.Resp(), c.Req())
			return
		}
	}
}

// 中间件包装器
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

// NotFoundHandler 全局404处理方法， 如果需要修改
func NotFoundHandler(c Ctx) error {
	return ErrNotFound
}

// MethodNotAllowedHandler 全局405处理方法
func MethodNotAllowedHandler(c Ctx) error {
	return ErrMethodNotAllowed
}

// 获取handler的名称
func HandlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

func M(m MiddlewareFunc) MiddlewareFunc {
	return m
}

func H(h HandlerFunc) HandlerFunc {
	return h
}

// HelloTwig! ~~
func HelloTwig(c Ctx) error {
	return c.Stringf(http.StatusOK, "Hello %s!", "Twig")
}

// Static 处理静态文件的HandlerFunc
func Static(r string) HandlerFunc {
	root := path.Clean(r)
	return func(c Ctx) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}
		name := filepath.Join(root, path.Clean("/"+p)) // 安全考虑 + "/"
		return c.File(name)
	}
}
