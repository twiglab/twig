package twig

import (
	"context"
	"net/http"
	"os"
	"sync"
)

const Version = "0.7.1.dev"

type M map[string]interface{}

// Identifier 标识符接口
type Identifier interface {
	ID() string
	Name() string
	Type() string
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
	Worker Worker // Worker 负责Http请求处理

	//ebus Notifier

	Debug bool

	pre []MiddlewareFunc
	mid []MiddlewareFunc

	pool sync.Pool

	plugins map[string]Plugin

	name string
	id   string
	typ  string
}

// 创建空的Twig
func TODO() *Twig {
	t := &Twig{
		Debug: false,

		name: "main",
		typ:  TwigName,

		plugins: make(map[string]Plugin),
		//ebus:    newbox(),
	}

	idGen := uuidGen{}
	t.id = idGen.NextID()
	t.UsePlugin(idGen)

	t.
		WithHttpErrorHandler(DefaultHttpErrorHandler).
		WithLogger(newLog(os.Stdout, "twig-")).
		WithMuxer(NewRadixTree()).
		WithWorker(NewWork())

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
	m.Attach(t)
	return t
}

func (t *Twig) WithWorker(w Worker) *Twig {
	t.Worker = w
	w.Attach(t)
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
func (t *Twig) UsePlugin(plugins ...Plugin) *Twig {
	for _, plugin := range plugins {
		Attach(plugin, t)
		t.plugins[plugin.ID()] = plugin
	}

	return t
}

// Plugin 获取Plugin
func (t *Twig) Plugin(id string) (p Plugin, ok bool) {
	p, ok = t.plugins[id]
	return
}

// ServeHTTP 实现`http.Handler`接口
func (t *Twig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := t.Muxer.Lookup(r.Method, GetReqPath(r), r)

	mc := c.(muxerCtx)
	mc.reset(w, r)

	defer mc.Release()

	h := Merge(func(ctx Ctx) error { //注意这里是个闭包，闭包中处理Twig级中间件，结束后处理Pre中间件
		handler := Merge(mc.Handler(), t.mid) // 处理Twig级中间件
		return handler(ctx)
	}, t.pre) // 处理Pre中间件

	if err := h(c); err != nil { // 链式调用，如果出错，交给Twig的HttpErrorHandler处理
		t.HttpErrorHandler(err, c)
	}
}

// Start Cycler#Start
func (t *Twig) Start() error {
	t.Logger.Printf("Twig@%s(id = %s ver = %s)\n", t.Name(), t.ID(), Version)

	for _, p := range t.plugins {
		if cycler, ok := p.(Cycler); ok {
			if err := cycler.Start(); err != nil {
				t.Logger.Printf("Plugin (id = %s) start fatal, Err = %v\n", p.ID(), err)
			}
		}
	}

	return t.Worker.Start()
}

// Start Cycler#Shutdown
func (t *Twig) Shutdown(ctx context.Context) error {
	for _, p := range t.plugins {
		if cycler, ok := p.(Cycler); ok {
			if err := cycler.Shutdown(ctx); err != nil {
				t.Logger.Printf("Plugin (id = %s) shoutdown fatal, Err = %v\n", p.ID(), err)
			}
		}
	}
	return t.Worker.Shutdown(ctx)
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
func (t *Twig) ID() (id string) {
	return t.id
}

func (t *Twig) Type() string {
	return t.typ
}

func (t *Twig) SetType(typ string) {
	t.typ = typ
}

func (t *Twig) Config() *Cfg {
	return Config(t).WithNamer(t)
}
