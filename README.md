# twig
twig 是一个面向webapi的简单的webserver 

> Twig 采用QQ群提供技术支持，QQ群号：472900117

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

Twig 是一个仔细设计过的webserver， 与其他的webserver不同，Twig的设计的目标是成为 `构建应用程序的基石`

Twig 的设计分为，核心，外围， 工具三个部分

## 核心

Twig 的核心组件包括：请求执行环境，服务器与连接器，日志，请求处理中间件，以及Twig本身

连接器负和服务器完成对网络协议处理构成请求，请求处理中间件负责对请求过滤执行，执行环境用于提供应用执行所需要的上下文，用于业务处理，Twig把所有的组建继承成为一个完整的应用

### 服务器与连接器

（coming soon...）

### 请求执行环境

请求执行环境的功能是为应用提供一个上下文（Ctx），有下列组件构成：

- Lookuper（路由执行器）用于查找符合当前请求路径的handler，并返回执行环境Ctx
- Matcher （匹配器）用于根据当前请求返回符合条件的路由执行器
- Wrapper（包装器）一个可以配置的环境包装器
- Ctx （执行上下文）提供请求上下文

除此之外，执行环境还包括：

- Register（注册器）提供handler注册功能，可以用Config工具进行配置
- Muxer（路由器）描述接口
- HandlerFunc（请求处理）

Twig 通过包装器的Match方法查找符合条件的路由查找器（应当为Muxer的实现），提供路由查找器的Lookup方法查找并执行路由，返回Ctx，用于执行Handler

```go
// 这个例子详细的演示了上述组建的组合使用场景
// 实现子域名，每个域名一个路由器
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/twiglab/twig"
)

type WWWMux struct {
	twig.Muxer
}

func (m *WWWMux) Match(r *http.Request) twig.Lookuper {
	if r.Host == "www.twiglab.org" {
		return m
	}
	return nil
}

func main() {

	www := &WWWMux{
		twig.NewRadixTree(),
	}

	other := twig.NewRadixTree()

	twig.NewConfig(www).
		Get("/", func(c twig.Ctx) error {
			return c.String(http.StatusOK, "Hello www.twiglab.org!")
		})

	twig.NewConfig(other).
		Get("/", func(c twig.Ctx) error {
			return c.String(http.StatusOK, "Hello twig.twiglab.org!")
		})

	web := twig.TODO()
	web.WithWrapper(twig.TwoMux(other, www))

	web.Start()

	twig.Signal(twig.Graceful(web, 15*time.Second), os.Interrupt)
}
```

通过实现不同的匹配器，可以实现在不同的地址，端口，http header，区分处理

Twig 提供了`MutiMux`工具辅助多路由集成，NewMutiMux 和 TwoMux 工具函数用于创建多路由和2个路由的Wrapper

静态文件处理并不是twig的擅长，提供多路由集成主要用于下来2个场景：

1. 用于在不同的地址，端口上暴露不同的服务，例如在特殊的地址上暴露监控服务，用于和主体应用隔离
2. subdomains

多余2个路由不常见，大多数情况下一个足以。twig默认的路由`RadixTree`直接实现了执行环境的所有接口，作为默认组建在Twig创建是默认加入

#### 为什么需要这么复杂？

基于以下考虑，twig的执行环境设计的比较复杂：

- twig支持多路由，即twig并没有限制只能是用RedixTree，只要实现Muxer接口，均可以作为路由使用，用于实现特殊的路由处理规则
- 现代应用都是在云环境下运行，作为一个应用的一部分（微服务），采集，上报，监控是构建应用的基本功能，twig有必要支持在不同的路由上提供不同的功能，做到功能隔离

再次说明：一般情况下，twig默认的RedixTree足以应付大部分场景

（阅读代码提示：RedixTree的实现位于redix.go中，路由处理需要考虑更快的执行效率，所以代码阅读起来可能有一些不太好理解）

### 中间件

（coming soon...）
