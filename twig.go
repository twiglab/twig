package twig

import (
	"context"
	"net/http"
	"os"
	"sync"
)

const Version = "0.8.4-dev"

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
	// 错误处理Handler
	HttpErrorHandler HttpErrorHandler

	Logger Logger // Logger 组件负责日志输出
	Server Server // Server 负责Http请求处理
	Muxer  Muxer  // 路由器

	Debug bool

	pre []MiddlewareFunc
	mid []MiddlewareFunc

	pool sync.Pool

	plugins map[string]Plugger

	name string
	id   string
	typ  string
}

// TODO 创建默认的Twig，后续可以用With*方法设置组建
func TODO() *Twig {
	t := &Twig{
		Debug: false,

		name: "main",
		typ:  "Twig",

		plugins: make(map[string]Plugger),

		HttpErrorHandler: DefaultHttpErrorHandler,
	}

	/*
		加入默认的UUID插件，twig的ID和RequestID中间件需要使用UUID插件
	*/

	idGen := uuidGen{}
	t.id = idGen.NextID()
	t.UsePlugger(idGen, &defaultBinder{})

	/*
		设置默认的Twig组建
	*/
	t.
		WithLogger(NewLog(os.Stdout, "twig-")).
		WithServer(NewWork()).
		WithMuxer(NewRadixTree())

	return t
}

func (t *Twig) WithMuxer(mux Muxer) *Twig {
	t.Muxer = mux
	return t
}

func (t *Twig) WithLogger(l Logger) *Twig {
	t.Logger = l
	return t
}

func (t *Twig) WithServer(w Server) *Twig {
	t.Server = w
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
func (t *Twig) UsePlugger(plugins ...Plugger) *Twig {
	for _, plugger := range plugins {
		Attach(plugger, t)
		t.plugins[plugger.ID()] = plugger
	}

	return t
}

// Plugin 获取Plugin
func (t *Twig) GetPlugger(id string) (p Plugger, ok bool) {
	p, ok = t.plugins[id]
	return
}

// ServeHTTP 实现`http.Handler`接口
func (t *Twig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, path := r.Method, GetReqPath(r)
	c := t.Muxer.Lookup(method, path, r)

	mc := c.(muxerCtx)
	mc.reset(w, r, t)

	defer mc.Release()

	h := Merge(func(ctx Ctx) error { //闭包，处理Twig级中间件，结束后处理Pre中间件
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

	return t.Server.Start()
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
	return t.Server.Shutdown(ctx)
}

// SetName Namer#SetName
func (t *Twig) SetName(name string) {
	t.name = name
}

// Name Identifier#Name
func (t *Twig) Name() string {
	return t.name
}

// ID Identifier#ID
func (t *Twig) ID() (id string) {
	return t.id
}

// Type Identifier#Type
func (t *Twig) Type() string {
	return t.typ
}

// SetType 设置当前应用程序类型
func (t *Twig) SetType(typ string) {
	t.typ = typ
}

// Config Configer#Config
func (t *Twig) Config() *Config {
	return NewConfig(t.Muxer)
}
