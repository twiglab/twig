# twig
twig 是一个面向webapi的简单的webserver 

> Twig 采用QQ群提供技术支持，QQ群号：472900117

## 安装

go get github.com/twiglab/twig

## 入门

```go
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/twiglab/twig"
)

func main() {
	web := twig.TODO()

	web.Config().
		Get("/hello", func(c twig.Ctx) error {
			return c.String(http.StatusOK, "Hello Twig!")
		})

	web.Start()

	twig.Signal(twig.Graceful(web, 15*time.Second), os.Interrupt)
}
```

- Twig的默认监听端口是4321, 可以通过twig.DefaultAddress全局变量修改(`位于var.go中`)，或者自定义自己的Server
- 使用twig.TODO()创建`默认的`Twig，默认的Twig包括默认的HttpServer（Work）,默认的路由实现（RadixTree)，默认的Logger和默认的HttpErrorHandler
- twig.Config是Twig提供的配置工具（注意：在Twig的世界里，工具的定义为附属品，并不是Twig必须的一部分），Twig没有像别的webserver一样提供GET，POST等方法，所有的配置工作都通过Config完成
- Twig要求所有的Server的实现必须是*非堵塞*的，Start方法将启动Twig，Twig提供了Signal组件用于堵塞应用，处理系统信号，完成和shell的交互


Twig最大的特点是简洁，灵活，Twig的所有组建都以接口方式提供，支持重写，Twig也提供了Plugger模块，集成其他组建，用于增强Twig的功能


至此讲述的内容，已经足够让您运行并使用Twig。 *祝您使用Twig愉快！*

----

## Twig的结构

Twig 是一个仔细设计过的webserver， 与其他的webserver不同，Twig的设计的目标是更好的做好一件事情。

Twig 的设计分为，核心，外围， 工具三个部分，核心是Twig的必须部分，外围不是必须的，但是外围可以更好的扩充Twig的功能，工具并不是Twig的一部分，Twig也不依赖任何工具，工具可以让Twig使用更加方便。

---


## 核心

Twig 的核心组件包括：路由器（由Muxer接口定义`mux.go`），服务器（由于Server接口定义`serve.go`），日志（由Logger接口定义`log.go`），上下文和处理函数和中间件(`ctx.go和handler.go`中定义)，以及Twig本身（`twig.go`）

服务器（Worker）的作用是处理HTTP链接，路由器（Muxer）的作用是找到（Lookup）指定的处理器HandlerFunc，处理器则是执行具体业务的地方，它通过上下文（Ctx）和Server交互，Twig负责将上述几个组件组合起来，形成一个Webserver

### 路由器（Muxer）
