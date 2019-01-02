# twig
twig is a simple web server （预览发布），API没有冻结，请注意

## 入门

```go
package main

import (
	"github.com/twiglab/twig"
	"github.com/twiglab/twig/middleware"
)

func main() {
	api := twig.Default()

	twig.Config(api).
		Use(middleware.Recover()).
		Get("/hello", twig.HelloTwig).
		Done()

	api.Start()

	twig.Signal(twig.Quit())
}
```
- Twig的默认监听端口是4321, 可以通过twig.DefaultAddress全局变量修改(`位于var.go中`)，或者自定义自己的Server
- 使用twig.Default()创建`默认的`Twig，默认的Twig包括默认的HttpServer（DefaultServnat）,默认的路由实现（RadixTree)，默认的Logger和默认的HttpErrorHandler
