package twig

// 组装器
type Assembler interface {
	Register
	PluginHelper
}

type target struct {
	Register
	PluginHelper
}

func newTarget(r Register, twig *Twig) Assembler {
	return &target{
		Register:     r,
		PluginHelper: twig,
	}
}

// Mouter 接口用于模块化设置路由
type Mounter interface {
	Mount(Assembler)
}

type MountFunc func(Assembler)

func (m MountFunc) Mount(target Assembler) {
	m(target)
}

// Conf Twig路由配置工具
type Conf struct {
	target Assembler
}

func config(r Register, twig *Twig) *Conf {
	return &Conf{
		target: newTarget(r, twig),
	}
}

func Config(r Register) *Conf {
	return &Conf{
		target: newTarget(r, nil),
	}
}

// Use 当前Register增加中间件
func (c *Conf) Use(m ...MiddlewareFunc) *Conf {
	c.Use(m...)
	return c
}

// AddHandler 增加Handler
func (c *Conf) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	c.target.AddHandler(method, path, handler, m...)
	return c
}

func (c *Conf) Get(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(GET, path, handler, m...)
}

func (c *Conf) Post(path string, handler HandlerFunc, m ...MiddlewareFunc) *Conf {
	return c.AddHandler(POST, path, handler, m...)
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

// Mount 挂载Mounter到当前Assembler
func (c *Conf) Mount(mount Mounter) *Conf {
	mount.Mount(c.target)
	return c
}

// Static 增加静态路由
func (c *Conf) Static(path, file string, m ...MiddlewareFunc) *Conf {
	return c.Get(path, Static(file), m...)
}

// Group 配置路由组
func (c *Conf) Group(path string, f MountFunc) *Conf {
	f(NewGroup(c.target, path))
	return c
}

/*
	web.Conf().
		Group("/api", func(r twig.Assembler) {
			twig.Config(r).
				Post("/addUser", func(c twig.Ctx) error {
					...
				})
		})
*/
// Group 提供路由分组支持
type Group struct {
	prefix string
	m      []MiddlewareFunc
	Assembler
}

func NewGroup(assembler Assembler, prefix string) *Group {
	return &Group{
		prefix:    prefix,
		Assembler: assembler,
	}
}

func (g *Group) Use(mid ...MiddlewareFunc) {
	g.m = append(g.m, mid...)
}

func (g *Group) AddHandler(method, path string, h HandlerFunc, m ...MiddlewareFunc) {
	handler := Merge(h, g.m)
	g.Assembler.AddHandler(method, g.prefix+path, handler, m...)
}
