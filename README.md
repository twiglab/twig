# twig
twig 是一个面向webapi的简单的webserver，Twig的目标是成为构建Golang微服务的基石

> Twig 采用QQ群提供技术支持，QQ群号：472900117

## 性能测试

```
goos: linux
goarch: amd64
pkg: web
BenchmarkTwigStatic-2      	   20000	     71232 ns/op	     870 B/op	       0 allocs/op
BenchmarkTwigGitHubAPI-2   	   10000	    101092 ns/op	     878 B/op	       0 allocs/op
BenchmarkTwigGplusAPI-2    	  300000	      5613 ns/op	      57 B/op	       0 allocs/op
BenchmarkTwigParseAPI-2    	  200000	     10660 ns/op	     173 B/op	       0 allocs/op

BenchmarkEchoStatic-2      	   20000	     63527 ns/op	    2126 B/op	     157 allocs/op
BenchmarkEchoGitHubAPI-2   	   20000	     91821 ns/op	    2496 B/op	     203 allocs/op
BenchmarkEchoGplusAPI-2    	  300000	      5072 ns/op	     161 B/op	      13 allocs/op
BenchmarkEchoParseAPI-2    	  200000	      9195 ns/op	     381 B/op	      26 allocs/op

BenchmarkGinStatic-2       	   20000	     75786 ns/op	    8405 B/op	     157 allocs/op
BenchmarkGinGitHubAPI-2    	   10000	    103458 ns/op	   10619 B/op	     203 allocs/op
BenchmarkGinGplusAPI-2     	  200000	      6049 ns/op	     710 B/op	      13 allocs/op
BenchmarkGinParseAPI-2     	  200000	     11559 ns/op	    1421 B/op	      26 allocs/op
PASS
ok  	web	24.538s
```

*Twig 比Gin更快，更好用!*

## 安装

go get github.com/twiglab/twig

*Twig 支持 go mod*

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
	web.AddServer(twig.NewServer(":4321")) // twig支持多server

    // twig支持多路由， 在默认路由上增加handler
	web.Config().
		Get("/hello", func(c twig.Ctx) error {
			return c.String(twig.OK, "Hello Twig!")
		})

	web.Start()

	twig.Signal(twig.Graceful(web, 15*time.Second), os.Interrupt)
}
```

- Twig的默认监听端口是4321, 或者自定义自己的Server
- 使用twig.TODO()创建`默认的`Twig，默认的Twig包括，默认的路由实现（RadixTree)，默认的Logger和默认的HttpErrorHandler
- twig.Config是Twig提供的配置工具，Twig没有像别的webserver一样提供GET，POST等方法，所有的配置工作都通过Config完成
- Twig要求所有的Server的实现必须是*非堵塞*的，Start方法将启动Twig，Twig提供了Signal组件用于堵塞应用，处理系统信号，完成和shell的交互


Twig最大的特点是简洁，灵活，Twig的所有组建都以接口方式提供，支持重写，Twig也提供了Plugger模块，集成其他组建，用于增强Twig的功能


至此讲述的内容，已经足够让您运行并使用Twig。 *祝您使用Twig愉快！*

----

## Twig的结构

Twig 是一个仔细设计过的webserver， 与其他的webserver不同，Twig的设计的目标是成为 `构建应用程序的基石`

Twig 的设计分为，核心，外围， 工具三个部分

## 核心

Twig 的核心组件包括：请求执行环境，服务器与连接器，日志，请求处理中间件，以及Twig本身

连接器和服务器完成对网络协议处理构成请求，请求处理中间件负责对请求过滤执行，执行环境用于提供应用执行所需要的上下文，用于业务处理，Twig把所有的组建继承成为一个完整的应用

### 服务器与连接器

（coming soon...）

### 请求执行环境

请求执行环境的功能是为应用提供一个上下文（Ctx），有下列组件构成：

- Lookuper（路由执行器）用于查找符合当前请求路径的handler，并返回执行环境Ctx
- Ctx （执行上下文）提供请求上下文

除此之外，执行环境还包括：

- Register（注册器）提供handler注册功能，可以用Config工具进行配置
- Muxer（路由器）描述接口
- HandlerFunc（请求处理）

执行环境的核心是Lookuper和Register 用于路由查找和路由注册（即Muxer接口）。Twig 通过路由查找器的Lookup方法查找并执行路由，返回Ctx，用于执行Handler


---

Ctx和HandlerFunc

### 中间件


（coming soon...）

## 插件 （Plugin）

## 多Server支持

