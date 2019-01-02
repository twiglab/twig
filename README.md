# twig
twig is a simple web server （预览发布），API没有冻结，请注意

## 简单使用

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
Twig的默认监听端口是4321, 可以通过twig.DefaultAddress全局变量修改(`位于var.go中`)，或者自定义自己的Server
