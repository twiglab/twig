package twig

import (
	"net/http"
	"net/url"
	"path"
	"path/filepath"
)

// HandlerFunc Twig的Handler方法
type HandlerFunc func(Ctx) error

type MiddlewareFunc func(HandlerFunc) HandlerFunc

// WrapHttpHandler 包装http.Handler 为HandlerFunc
func WrapHttpHandler(h http.Handler) HandlerFunc {
	return func(c Ctx) error {
		h.ServeHTTP(c.Resp(), c.Req())
		return nil
	}
}

// Merge 中间件包装器
func Merge(handler HandlerFunc, m []MiddlewareFunc) HandlerFunc {
	if m == nil {
		return handler
	}
	h := handler
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// NotFoundHandler 全局404处理方法
var NotFoundHandler = func(c Ctx) error {
	return ErrNotFound
}

// MethodNotAllowedHandler 全局405处理方法
var MethodNotAllowedHandler = func(c Ctx) error {
	return ErrMethodNotAllowed
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

// ServerInfo ServerInfo 中间件将Twig#Name()设置 Server 头
// Debug状态下，返回 x-powerd-by 为Twig#Type()
func ServerInfo() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Ctx) error {
			w := c.Resp()
			w.Header().Set(HeaderServer, c.Twig().Name())
			if c.Twig().Debug {
				w.Header().Set(HeaderXPoweredBy, c.Twig().Type())
			}
			return next(c)
		}
	}
}
