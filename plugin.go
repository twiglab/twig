package twig

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
// Binder 作为一个插件集成到Twig中
type Binder interface {
	Bind(interface{}, Ctx) error
}

// GetBinder 获取绑定接口
func GetBinder(id string, c Ctx) (Binder, bool) {
	plugin, ok := GetPlugin(id, c)
	if !ok {
		return nil, false
	}
	if binder, ok := plugin.(Binder); ok {
		return binder, ok
	}
	return nil, false
}
