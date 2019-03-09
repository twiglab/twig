package twig

import (
	"io"

	"github.com/twiglab/twig/internal/uuid"
)

// Plugger 定义了Twig的外部插件
// 如果插件需要生命周期管理，请实现Cycler接口
// 如果插件需要访问Twig本身，请实现Attacher接口
type Plugger interface {
	Identifier
}

type PluginHelper interface {
	UsePlugger(...Plugger)
	GetPlugger(string) Plugger
}

// GetPlugger 从当前Ctx中获取Plugger
func GetPlugger(id string, c Ctx) Plugger {
	t := c.Twig()
	return t.GetPlugger(id)
}

// Binder 数据绑定接口
// Binder 作为一个插件集成到Twig中,请实现Plugger接口
type Binder interface {
	Bind(interface{}, Ctx) error
}

// GetBinder 获取绑定接口
func GetBinder(id string, c Ctx) Binder {
	return GetPlugger(id, c).(Binder)
}

type Renderer interface {
	Render(io.Writer, string, interface{}, Ctx) error
}

func GetRenderer(id string, c Ctx) Renderer {
	return GetPlugger(id, c).(Renderer)
}

// IdGenerator ID发生器接口
type IdGenerator interface {
	NextID() string
}

func GetIdGenerator(id string, c Ctx) IdGenerator {
	plugger := GetPlugger(id, c)
	return plugger.(IdGenerator)
}

const uuidPluginID = "_twig_uuid_plugin_id_"

type uuidGen struct {
}

func (id uuidGen) ID() string {
	return uuidPluginID
}

func (id uuidGen) NextID() string {
	return uuid.NewV1().String32()
}

func GenID(c Ctx) string {
	idgen := GetIdGenerator(uuidPluginID, c)
	return idgen.NextID()
}
