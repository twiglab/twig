package twig

import (
	"io"

	"github.com/twiglab/twig/internal/uuid"
)

// Plugin 定义了Twig的外部插件
// 如果插件需要生命周期管理，请实现Cycler接口
// 如果插件需要访问Twig本身，请实现Attacher接口
type Plugin interface {
	Identifier
}

// GetPlugin 从当前Ctx中获取Plugin
func GetPlugin(id string, c Ctx) (p Plugin, ok bool) {
	t := c.Twig()
	p, ok = t.Plugin(id)
	return
}

// UsePlugin 将plugin加入到Twig中
func UsePlugin(t *Twig, plugin ...Plugin) {
	t.UsePlugin(plugin...)
}

// Binder 数据绑定接口
// Binder 作为一个插件集成到Twig中,请实现Plugin接口
type Binder interface {
	Bind(interface{}, Ctx) error
}

// GetBinder 获取绑定接口
func GetBinder(id string, c Ctx) (binder Binder, ok bool) {
	var plugin Plugin
	if plugin, ok = GetPlugin(id, c); ok {
		binder, ok = plugin.(Binder)
	}
	return
}

type Renderer interface {
	Render(io.Writer, string, interface{}, Ctx) error
}

func GetRenderer(id string, c Ctx) (r Renderer, ok bool) {
	var plugin Plugin
	if plugin, ok = GetPlugin(id, c); ok {
		r, ok = plugin.(Renderer)
	}
	return
}

type IdGenerator interface {
	NextID() string
}

func GetIdGenerator(id string, c Ctx) (gen IdGenerator, ok bool) {
	var plugin Plugin
	if plugin, ok = GetPlugin(id, c); ok {
		gen, ok = plugin.(IdGenerator)
	}
	return
}

const uuidPluginID = "_twig_uuid_plugin_id_"

type uuidGen struct {
}

func (id uuidGen) ID() string {
	return uuidPluginID
}

func (id uuidGen) Name() string {
	return uuidPluginID
}

func (id uuidGen) Type() string {
	return "idGen"
}

func (id uuidGen) NextID() string {
	return uuid.NewV1().String()
}

func GetUUIDGen(c Ctx) (gen IdGenerator) {
	gen, _ = GetIdGenerator(uuidPluginID, c)
	return
}
