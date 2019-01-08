package twig

import (
	"context"
	"net/http"
	"os"
	"sync"
)

type Map map[string]interface{}

// Identifier 标识符接口
type Identifier interface {
	ID() string
	Name() string
}

// Attacher 用于设置Twig和组件之间的联系
type Attacher interface {
	Attach(*Twig)
}

// Cycler 设置周期管理
type Cycler interface {
	Start() error
	Shutdown(context.Context) error
}

// Namer 命名接口
type Namer interface {
	SetName(string)
}

// Twig
type Twig struct {
	HttpErrorHandler HttpErrorHandler

	Logger Logger // Logger 组件负责日志输出
	Muxer  Muxer  // Muxer 组件负责路由处理
	Server Server // Server 负责Http请求处理

	Debug bool

	pre []MiddlewareFunc
	mid []MiddlewareFunc

	pool sync.Pool

	plugins map[string]Plugin

	name string
}

// 创建空的Twig
func TODO() *Twig {
	t := &Twig{
		Debug:   false,
		name:    "main",
		plugins: make(map[string]Plugin),
	}
	t.pool.New = func() interface{} {
		return t.NewCtx(nil, nil)
	}

	t.
		WithServer(DefaultServant()).
		WithHttpErrorHandler(DefaultHttpErrorHandler).
		WithLogger(newLog(os.Stdout, "twig-log-")).
		WithMuxer(NewRadixTree())

	return t
}

// 创建默认的Twig
func Default() *Twig {
	t := TODO()
	t.UsePlugin(&DefaultBinder{})
	return t
}

func (t *Twig) WithLogger(l Logger) *Twig {
	t.Logger = l
	Attach(l, t)
	return t
}

func (t *Twig) WithHttpErrorHandler(eh HttpErrorHandler) *Twig {
	t.HttpErrorHandler = eh
	return t
}

func (t *Twig) WithMuxer(m Muxer) *Twig {
	t.Muxer = m
	Attach(m, t)
	return t
}

func (t *Twig) WithServer(s Server) *Twig {
	t.Server = s
	s.Attach(t)
	return t
}

func (t *Twig) EnableDebug() *Twig {
	t.Debug = true
	return t
}

// Pre 中间件支持， 注意Pre中间件工作在路由之前
func (t *Twig) Pre(m ...MiddlewareFunc) {
	t.pre = append(t.pre, m...)
}

// Use Twig级中间件支持
func (t *Twig) Use(m ...MiddlewareFunc) {
	t.mid = append(t.mid, m...)
}

// UserPlugin 加入Plugin
func (t *Twig) UsePlugin(plugins ...Plugin) {
	for _, plugin := range plugins {
		Attach(plugin, t)
		t.plugins[plugin.ID()] = plugin
	}
}

// Plugin 获取Plugin
func (t *Twig) Plugin(id string) (p Plugin, ok bool) {
	p, ok = t.plugins[id]
	return
}

// ServeHTTP 实现`http.Handler`接口
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

// Start Cycler#Start
func (t *Twig) Start() error {
	t.Logger.Printf(banner, Version)

	for _, p := range t.plugins {
		Start(p)
	}

	return t.Server.Start()
}

// Start Cycler#Shutdown
func (t *Twig) Shutdown(ctx context.Context) error {
	for _, p := range t.plugins {
		Shutdown(p, ctx)
	}

	return t.Server.Shutdown(ctx)
}

// 面向第三方路由，提供Ctx的创建功能
// 注意：Twig 不管理第三方路由使用的Ctx，只负责创建，不负责回收
func (t *Twig) NewCtx(w http.ResponseWriter, r *http.Request) Ctx {
	return &ctx{
		req:     r,
		resp:    NewResponseWarp(w),
		t:       t,
		store:   make(Map),
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

// SetName Namer#SetName
func (t *Twig) SetName(name string) {
	t.name = name
}

// Name Identifier#Name
func (t *Twig) Name() string {
	return t.name
}

// Name Identifier#ID
func (t *Twig) ID() string {
	return "Twig@" + t.name
}

func (t *Twig) Config() *Cfg {
	return Config(t).WithNamer(t)
}
