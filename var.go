package twig

// MaxParam URL中最大的参数，注意这个是全局生效的，
// 在自定义路由实现中，应该调整这个值到参数个数
// Twig自带的RadixTree路由会自动调整MaxParam
// 参考Twig#NewCtx (`twig.go`)
var MaxParam int = 5

// DefaultAddress 为Twig默认服务端口
// 只是影响Twig自带的DefaultServant
// 自定义的Server不受此变量影响
var DefaultAddress = ":4321"
