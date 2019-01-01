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
    api := twig.TODO()
    api.Use(middleware.Recover())
    api.Get("/hello", twig.HelloTwig)
    api.Start()
}

```
