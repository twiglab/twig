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
