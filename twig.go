package twig

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type M map[string]interface{}

type Cycler interface {
	Start() error
	Shutdown(context.Context) error
}

type Attacher interface {
	Attach(*Twig)
}

type Nameder interface {
	SetName(string)
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

	parteners map[string]Partner

	name string
}

// 创建默认的Twig
func TODO() *Twig {
	t := &Twig{
		Debug: false,
		name:  "main",
	}
	t.pool.New = func() interface{} {
		return t.NewCtx(nil, nil)
	}
	t.WithServer(DefaultServnat()).
		WithHttpErrorHandler(DefaultHttpErrorHandler).
		WithLogger(newLog(os.Stdout, "twig-log-")).
		WithMuxer(NewRadixTree())
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
func (t *Twig) Pre(m ...MiddlewareFunc) {
	t.pre = append(t.pre, m...)
}

// Twig级中间件支持
func (t *Twig) Use(m ...MiddlewareFunc) {
	t.mid = append(t.mid, m...)
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

func (t *Twig) AddPartner(ps ...Partner) *Twig {
	for _, p := range ps {
		attach(p, t)
		t.parteners[p.ID()] = p
	}
	return t
}

func (t *Twig) Partner(id string) (p Partner, ok bool) {
	p, ok = t.parteners[id]
	return
}

// 实现http.Handler
func (t *Twig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := t.pool.Get().(*ctx) // pool 中获取Ctx
	c.Reset(w, r)            // 重置Ctx，放入当前的Resp和Req , Ctx是可以重用的

	h := Enhance(func(ctx Ctx) error { //注意这里是个闭包，闭包中处理Twig级中间件，结束后处理Pre中间件
		t.Muxer.Lookup(r.Method, GetReqPath(r), r, c) // 路由对当前Ctx实现装配
		handler := Enhance(c.Handler(), t.mid)        // 处理Twig级中间件
		return handler(ctx)
	}, t.pre)

	if err := h(c); err != nil { // 链式调用，如果出错，交给Twig的HttpErrorHandler处理
		t.HttpErrorHandler(err, c)
	}

	t.pool.Put(c) // 交还Ctx，后续复用，Http处理过程结束
}

func (t *Twig) Start() error {
	t.Logger.Println(banner)

	for _, p := range t.parteners {
		if cycler, ok := p.(Cycler); ok {
			cycler.Start()
		}
	}

	return t.Server.Start()
}

func (t *Twig) Shutdown(ctx context.Context) error {
	for _, p := range t.parteners {
		if cycler, ok := p.(Cycler); ok {
			cycler.Shutdown(ctx)
		}
	}

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

func (t *Twig) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) Route {
	return t.Muxer.AddHandler(method, path, handler, m...)
}

func (t *Twig) SetName(name string) {
	t.name = name
}

func (t *Twig) Name() string {
	return t.name
}

func (t *Twig) ID() string {
	return fmt.Sprintf("Twig@%s", t.name)
}

func (t *Twig) Type() string {
	return "webserver"
}

func (t *Twig) Config() *Config {
	return Cfg().With(t, t)
}
