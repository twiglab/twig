package twig

import (
	"context"
	"net/http"
	"os"
	"sync"
)

type M map[string]interface{}

type Attacher interface {
	Attach(*Twig)
}

type Identifier interface {
	Name() string
	Type() string
	Desc() string
}

type Twig struct {
	HttpErrorHandler HttpErrorHandler

	Logger Logger
	Muxer  Muxer
	Server Server

	Debug bool

	pre []MiddlewareFunc
	mid []MiddlewareFunc

	pool sync.Pool
}

// 创建默认的Twig
func TODO() *Twig {
	t := &Twig{
		Debug: false,
	}
	t.pool.New = func() interface{} {
		return t.NewCtx(nil, nil)
	}
	t.WithServer(DefaultServnat()).
		WithHttpErrorHandler(DefaultHttpErrorHandler).
		WithLogger(newLog(os.Stdout, "twig-log-")).
		WithMuxer(NewRadixTreeMux())
	return t
}

func (t *Twig) WithLogger(l Logger) *Twig {
	t.Logger = l
	attach(l, t)
	return t
}

func (t *Twig) WithHttpErrorHandler(eh HttpErrorHandler) *Twig {
	t.HttpErrorHandler = eh
	return t
}

// Pre 中间件支持， 注意Pre中间件工作在路由之前
func (t *Twig) Pre(m ...MiddlewareFunc) *Twig {
	t.pre = append(t.pre, m...)
	return t
}

// Twig级中间件支持
func (t *Twig) Use(m ...MiddlewareFunc) *Twig {
	t.mid = append(t.mid, m...)
	return t
}

func (t *Twig) WithMuxer(m Muxer) *Twig {
	t.Muxer = m
	attach(m, t)
	return t
}

func (t *Twig) WithServer(s Server) *Twig {
	t.Server = s
	s.Attach(t)
	return t
}

// 实现http.Handler
func (t *Twig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := t.pool.Get().(*ctx) // pool 中获取Ctx
	c.Reset(w, r)            // 重置Ctx，放入当前的Resp和Req , Ctx是可以重用的

	h := Enhance(func(ctx Ctx) error { //注意这里是个闭包，闭包中处理Twig级中间件，结束后处理Pre中间件
		t.Muxer.Lookup(r.Method, GetReqPath(r), r, c) // 路由对当前Ctx实现装配
		handler := Enhance(c.Handler(), t.mid)        // 处理Twig级中间件
		return handler(c)
	}, t.pre)

	if err := h(c); err != nil { // 链式调用，如果出错，交给Twig的HttpErrorHandler处理
		t.HttpErrorHandler(err, c)
	}

	t.pool.Put(c) // 交还Ctx，后续复用，Http处理过程结束
}

func (t *Twig) Start() error {
	t.Logger.Println(banner)
	return t.Server.Start()
}

func (t *Twig) Shutdown(ctx context.Context) error {
	return t.Server.Shutdown(ctx)
}

// 面向第三方路由，提供Ctx的创建功能
// 注意：Twig 不管理第三方路由使用的Ctx，只创建，不回收
func (t *Twig) NewCtx(w http.ResponseWriter, r *http.Request) Ctx {
	return &ctx{
		req:     r,
		resp:    NewResponseWarp(w),
		t:       t,
		store:   make(M),
		pvalues: make([]string, MaxParam),
		handler: NotFoundHandler,
	}
}

func (t *Twig) AcquireCtx() Ctx {
	c := t.pool.Get().(*ctx)
	return c
}

func (t *Twig) ReleaseCtx(c Ctx) {
	t.pool.Put(c)
}

func (t *Twig) add(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.Muxer.Add(method, path, handler, m...)
}

func (t *Twig) Get(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(GET, path, handler, m...)
}

func (t *Twig) Post(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(POST, path, handler, m...)
}

func (t *Twig) Delete(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(DELETE, path, handler, m...)
}

func (t *Twig) Put(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(PUT, path, handler, m...)
}

func (t *Twig) Patch(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(PATCH, path, handler, m...)
}

func (t *Twig) Head(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(HEAD, path, handler, m...)
}

func (t *Twig) Options(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(OPTIONS, path, handler, m...)
}

func (t *Twig) Trace(path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return t.add(TRACE, path, handler, m...)
}
