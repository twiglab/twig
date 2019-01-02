package twig

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
)

type HttpErrorHandler func(error, Ctx)

type HandlerFunc func(Ctx) error
type MiddlewareFunc func(HandlerFunc) HandlerFunc

func NopMiddleware(h HandlerFunc) HandlerFunc {
	return h
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

// 默认的错误处理
func DefaultHttpErrorHandler(err error, c Ctx) {
	var code int = http.StatusInternalServerError
	var msg interface{}

	if e, ok := err.(*HttpError); ok {
		code = e.Code
		msg = e.Msg

		if e.Internal != nil {
			err = fmt.Errorf("%v, %v", err, e.Internal)
		}
	} else {
		msg = http.StatusText(code)
	}

	if m, ok := msg.(string); ok {
		msg = map[string]string{"msg": m}
	}

	if !c.Resp().Committed {
		if c.Req().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			c.Logger().Println(err)
		}
	}
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
