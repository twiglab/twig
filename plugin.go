package twig

// Plugin 定义了Twig的外部插件
// 如果插件需要生命周期管理，请实现Cycler接口
// 如果插件需要访问Twig本身，请实现Attacher接口
type Plugin interface {
	Identifier
	Name() string
}

// GetPlugin 从当前Ctx中获取Plugin
// 注意：Plugin可能为nil
func GetPlugin(id string, c Ctx) Plugin {
	t := c.Twig()
	return t.Plugin(id)
}
