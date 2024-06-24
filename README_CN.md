[English](./README.md) | 中文

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/orbit)](https://goreportcard.com/report/github.com/shengyanli1982/orbit)
[![Build Status](https://github.com/shengyanli1982/orbit/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/orbit/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/orbit.svg)](https://pkg.go.dev/github.com/shengyanli1982/orbit)

# 简介

`orbit` 是一个轻量级的 HTTP web 服务包装框架，设计上注重简单易用。它提供了一系列便利的功能，帮助您快速构建和维护 web 服务。

`orbit` 的名字反映了该框架的目标，即封装构建 web 服务的复杂性，让您能够专注于核心业务逻辑，就像卫星平稳地环绕地球轨道一样。

### 为什么不直接使用 `gin`？

虽然 `gin` 是一个优秀的框架，但它需要额外的设置来进行日志记录和监控。`orbit` 基于 `gin` 构建，提供了开箱即用的这些功能，简化了启动 web 服务的过程。

# 优点

-   轻量级和用户友好
-   支持 `zap` 日志记录，包括 `async` 和 `sync` 模式
-   集成 `prometheus` 进行监控
-   包含 `swagger` API 文档支持
-   优雅的服务器关闭
-   `cors` 中间件支持跨源请求
-   自动恢复 panic
-   可定制的中间件
-   灵活的路由组
-   可定制的访问日志格式和字段
-   支持重复读取请求/响应体和缓存

# 安装

```bash
go get github.com/shengyanli1982/orbit
```

# 快速开始

`orbit` 旨在快速简便地进行 web 服务开发。只需按照以下简单步骤操作：

1. 创建 `orbit` 配置。
2. 定义 `orbit` 功能选项。
3. 创建一个 `orbit` 实例。

**默认的 URL 路径**

> [!NOTE]
>
> 默认的 URL 路径是系统定义的，无法更改。

-   `/metrics` - Prometheus 指标
-   `/swagger/*any` - Swagger API 文档
-   `/debug/pprof/*any` - PProf 调试
-   `/ping` - 健康检查

## 1. 配置

`orbit` 提供了几个配置选项，可以在启动 `orbit` 实例之前进行设置。

-   `WithSugaredLogger` - 使用 `zap` sugared logger（默认值：`DefaultSugaredLogger`）。
-   `WithLogger` - 使用 `zap` logger（默认值：`DefaultConsoleLogger`）。
-   `WithAddress` - HTTP 服务器监听地址（默认值：`127.0.0.1`）。
-   `WithPort` - HTTP 服务器监听端口（默认值：`8080`）。
-   `WithRelease` - HTTP 服务器发布模式（默认值：`false`）。
-   `WithHttpReadTimeout` - HTTP 服务器读取超时时间（默认值：`15s`）。
-   `WithHttpWriteTimeout` - HTTP 服务器写入超时时间（默认值：`15s`）。
-   `WithHttpReadHeaderTimeout` - HTTP 服务器读取请求头超时时间（默认值：`15s`）。
-   `WithAccessLogEventFunc` - HTTP 服务器访问日志事件函数（默认值：`DefaultAccessEventFunc`）。
-   `WithRecoveryLogEventFunc` - HTTP 服务器恢复日志事件函数（默认值：`DefaultRecoveryEventFunc`）。
-   `WithPrometheusRegistry` - HTTP 服务器 Prometheus 注册器（默认值：`prometheus.DefaultRegister`）。

您可以使用 `NewConfig` 创建默认配置，并使用 `WithXXX` 方法设置配置选项。`DefaultConfig` 是 `NewConfig()` 的别名。

## 2. 功能

`orbit` 提供了几个功能选项，可以在启动 `orbit` 实例之前进行设置：

-   `EnablePProf` - 启用 pprof 调试（默认值：`false`）
-   `EnableSwagger` - 启用 Swagger API 文档（默认值：`false`）
-   `EnableMetric` - 启用 Prometheus 指标（默认值：`false`）
-   `EnableRedirectTrailingSlash` - 启用重定向尾部斜杠（默认值：`false`）
-   `EnableRedirectFixedPath` - 启用重定向固定路径（默认值：`false`）
-   `EnableForwardedByClientIp` - 启用通过客户端 IP 转发（默认值：`false`）
-   `EnableRecordRequestBody` - 启用记录请求体（默认值：`false`）

您可以使用 `NewOptions` 创建一个空功能，并使用 `EnableXXX` 方法设置功能选项。

-   `DebugOptions` 用于调试，是 `NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()` 的别名。
-   `ReleaseOptions` 用于发布，是 `NewOptions().EnableMetric()` 的别名。

> [!NOTE]
>
> 推荐在调试时使用 `DebugOptions`，在发布时使用 `ReleaseOptions`。

## 3. 创建实例

一旦您创建了 `orbit` 的配置和功能选项，就可以创建一个 `orbit` 实例。

> [!IMPORTANT]
>
> 运行 `orbit` 实例时，它不会阻塞当前的 goroutine。这意味着在运行 `orbit` 实例后，您可以继续做其他事情。
>
> 如果您想阻塞当前的 goroutine，可以使用项目 [`GS`](https://github.com/shengyanli1982/gs) 提供的 `Waitting` 函数来阻塞当前的 goroutine。

> [!TIP]
>
> 为了简化流程，您可以使用 `NewHttpService` 将 `func(*gin.RouterGroup)` 包装成 `Service` 接口的实现。
>
> ```go
> NewHttpService(func(g *gin.RouterGroup) {
>   g.GET("/demo", func(c *gin.Context) {
>       c.String(http.StatusOK, "demo")
>   })
> })
> ```

**示例**

```go
package main

import (
	"time"

	"github.com/shengyanli1982/orbit"
)

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
{"level":"INFO","time":"2024-01-10T17:00:13.139+0800","logger":"default","caller":"orbit/gin.go:160","message":"http server is ready","address":"127.0.0.1:8080"}
```

**测试**

```bash
$ curl -i http://127.0.0.1:8080/ping
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 10 Jan 2024 09:07:26 GMT
Content-Length: 7

successs
```

## 4. 自定义中间件

`orbit` 基于 `gin`，因此您可以直接使用 `gin` 的中间件。这使您能够为特定任务实现自定义中间件。例如，您可以使用 `demo` 中间件在控制台打印 `>>>>>>!!! demo`。

以下是在 `orbit` 中使用自定义中间件的示例：

**示例**

```go
package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
)

// customMiddleware 函数定义了一个自定义的中间件
// The customMiddleware function defines a custom middleware
func customMiddleware() gin.HandlerFunc {
	// 返回一个 Gin 的 HandlerFunc
	// Return a Gin HandlerFunc
	return func(c *gin.Context) {
		// 打印一条消息
		// Print a message
		fmt.Println(">>>>>>!!! demo")

		// 调用下一个中间件或处理函数
		// Call the next middleware or handler function
		c.Next()
	}
}

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {})
}

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项，并启用 metric
	// Create a new Orbit feature options and enable metric
	opts := orbit.NewOptions().EnableMetric()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的中间件
	// Register a custom middleware
	engine.RegisterMiddleware(customMiddleware())

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
$ go run demo.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] GET    /metrics                  --> github.com/shengyanli1982/orbit/utils/wrapper.WrapHandlerToGin.func1 (2 handlers)
[GIN-debug] GET    /demo                     --> main.(*service).RegisterGroup.func1 (7 handlers)
{"level":"INFO","time":"2024-01-10T20:03:38.869+0800","logger":"default","caller":"orbit/gin.go:162","message":"http server is ready","address":"127.0.0.1:8080"}
>>>>>>!!! demo
{"level":"INFO","time":"2024-01-10T20:03:41.275+0800","logger":"default","caller":"log/default.go:10","message":"http server access log","id":"","ip":"127.0.0.1","endpoint":"127.0.0.1:59787","path":"/demo","method":"GET","code":200,"status":"OK","latency":"780ns","agent":"curl/8.1.2","query":"","reqContentType":"","reqBody":""}
```

## 5. 自定义路由组

`orbit` 中的自定义路由组功能允许您为 `demo` 服务注册一个自定义的路由组。例如，您可以注册像 `/demo` 和 `/demo/test` 这样的路由。

**示例**

```go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	ocom "github.com/shengyanli1982/orbit/common"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 注册一个自定义的路由组 "/demo"
	// Register a custom router group "/demo"
	g = g.Group("/demo")

	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET(ocom.EmptyURLPath, func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})

	// 在 "/demo/test" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo/test" path
	g.GET("/test", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "test" 字符串
		// Return HTTP status code 200 and the string "test"
		c.String(http.StatusOK, "test")
	})
}

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项，并启用 metric
	// Create a new Orbit feature options and enable metric
	opts := orbit.NewOptions().EnableMetric()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
$ curl -i http://127.0.0.1:8080/demo
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: POST, GET, OPTIONS, PUT, DELETE, UPDATE
Access-Control-Allow-Origin: *
Access-Control-Expose-Headers: Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type
Access-Control-Max-Age: 172800
Content-Type: text/plain; charset=utf-8
Date: Wed, 10 Jan 2024 12:09:37 GMT
Content-Length: 4

demo

$ curl -i http://127.0.0.1:8080/demo/test
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: POST, GET, OPTIONS, PUT, DELETE, UPDATE
Access-Control-Allow-Origin: *
Access-Control-Expose-Headers: Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type
Access-Control-Max-Age: 172800
Content-Type: text/plain; charset=utf-8
Date: Wed, 10 Jan 2024 12:09:43 GMT
Content-Length: 4

test
```

## 6. 自定义访问日志

要自定义 `orbit` 中的访问日志格式和字段，您可以使用以下示例：

**默认的 LogEvent 字段**

```go
// LogEvent 结构体用于记录日志事件
// The LogEvent struct is used to log events
type LogEvent struct {
	// Message 字段表示日志消息
	// The Message field represents the log message
	Message string `json:"message,omitempty" yaml:"message,omitempty"`

	// ID 字段表示事件的唯一标识符
	// The ID field represents the unique identifier of the event
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// IP 字段表示发起请求的IP地址
	// The IP field represents the IP address of the request initiator
	IP string `json:"ip,omitempty" yaml:"ip,omitempty"`

	// EndPoint 字段表示请求的终端点
	// The EndPoint field represents the endpoint of the request
	EndPoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Path 字段表示请求的路径
	// The Path field represents the path of the request
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Method 字段表示请求的HTTP方法
	// The Method field represents the HTTP method of the request
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Code 字段表示响应的HTTP状态码
	// The Code field represents the HTTP status code of the response
	Code int `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`

	// Status 字段表示请求的状态
	// The Status field represents the status of the request
	Status string `json:"status,omitempty" yaml:"status,omitempty"`

	// Latency 字段表示请求的延迟时间
	// The Latency field represents the latency of the request
	Latency string `json:"latency,omitempty" yaml:"latency,omitempty"`

	// Agent 字段表示发起请求的用户代理
	// The Agent field represents the user agent of the request initiator
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`

	// ReqContentType 字段表示请求的内容类型
	// The ReqContentType field represents the content type of the request
	ReqContentType string `json:"reqContentType,omitempty" yaml:"reqContentType,omitempty"`

	// ReqQuery 字段表示请求的查询参数
	// The ReqQuery field represents the query parameters of the request
	ReqQuery string `json:"query,omitempty" yaml:"query,omitempty"`

	// ReqBody 字段表示请求的主体内容
	// The ReqBody field represents the body of the request
	ReqBody string `json:"reqBody,omitempty" yaml:"reqBody,omitempty"`

	// Error 字段表示请求中的任何错误
	// The Error field represents any errors in the request
	Error any `json:"error,omitempty" yaml:"error,omitempty"`

	// ErrorStack 字段表示错误的堆栈跟踪
	// The ErrorStack field represents the stack trace of the error
	ErrorStack string `json:"errorStack,omitempty" yaml:"errorStack,omitempty"`
}
```

**示例**

```go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})
}

// customAccessLogger 函数定义了一个自定义的访问日志记录器
// The customAccessLogger function defines a custom access logger
func customAccessLogger(logger *zap.SugaredLogger, event *log.LogEvent) {
	// 记录访问日志，包括路径和方法
	// Log the access, including the path and method
	logger.Infow("access log", "path", event.Path, "method", event.Method)
}

func main() {
	// 创建一个新的 Orbit 配置，并设置访问日志事件函数
	// Create a new Orbit configuration and set the access log event function
	config := orbit.NewConfig().WithAccessLogEventFunc(customAccessLogger)

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
{"level":"INFO","time":"2024-01-10T20:22:01.244+0800","logger":"default","caller":"accesslog/demo.go:24","message":"access log","path":"/demo","method":"GET"}
```

## 7. 自定义恢复日志

HTTP 服务器的恢复日志可以帮助您了解当服务遇到 panic 时发生了什么。使用 `orbit`，您可以自定义恢复日志的格式和字段。以下是如何自定义恢复日志格式和字段的示例：

**示例**

```go
package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数，该函数会触发 panic
	// Register a GET method handler function on the "/demo" path, this function will trigger a panic
	g.GET("/demo", func(c *gin.Context) {
		panic("demo")
	})
}

// customRecoveryLogger 函数定义了一个自定义的恢复日志记录器
// The customRecoveryLogger function defines a custom recovery logger
func customRecoveryLogger(logger *zap.SugaredLogger, event *log.LogEvent) {
	// 记录恢复日志，包括路径、方法、错误和错误堆栈
	// Log the recovery, including the path, method, error, and error stack
	logger.Infow("recovery log", "path", event.Path, "method", event.Method, "error", event.Error, "errorStack", event.ErrorStack)
}

func main() {
	// 创建一个新的 Orbit 配置，并设置恢复日志事件函数
	// Create a new Orbit configuration and set the recovery log event function
	config := orbit.NewConfig().WithRecoveryLogEventFunc(customRecoveryLogger)

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
{"level":"INFO","time":"2024-01-10T20:27:10.041+0800","logger":"default","caller":"recoverylog/demo.go:22","message":"recovery log","path":"/demo","method":"GET","error":"demo","errorStack":"goroutine 6 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x65\ngithub.com/shengyanli1982/orbit/internal/middleware.Recovery.func1.1()\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:145 +0x559\npanic({0x170ec80, 0x191cb70})\n\t/usr/local/go/src/runtime/panic.go:884 +0x213\nmain.(*service).RegisterGroup.func1(0x0?)\n\t/Volumes/DATA/programs/GolandProjects/orbit/example/recoverylog/demo.go:17 +0x27\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.AccessLogger.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:59 +0x1a5\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.Cors.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:35 +0x139\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.BodyBuffer.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/buffer.go:18 +0x92\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.Recovery.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:166 +0x82\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/gin-gonic/gin.(*Engine).handleHTTPRequest(0xc0000076c0, 0xc0001e6300)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/gin.go:616 +0x66b\ngithub.com/gin-gonic/gin.(*Engine).ServeHTTP(0xc0000076c0, {0x1924a30?, 0xc0000c02a0}, 0xc0001e6200)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/gin.go:572 +0x1dd\nnet/http.serverHandler.ServeHTTP({0xc00008be30?}, {0x1924a30, 0xc0000c02a0}, 0xc0001e6200)\n\t/usr/local/go/src/net/http/server.go:2936 +0x316\nnet/http.(*conn).serve(0xc0000962d0, {0x19253e0, 0xc00008bd40})\n\t/usr/local/go/src/net/http/server.go:1995 +0x612\ncreated by net/http.(*Server).Serve\n\t/usr/local/go/src/net/http/server.go:3089 +0x5ed\n"}
```

## 8. 异步日志

`orbit` 利用 `law` 项目提供一个 `async` 日志。以下是一个示例，展示如何在 `demo` 服务中启用 `async` 日志。

> [!TIP]
>
> [`law`](https://github.com/shengyanli1982/law) 是一个为 `zap`、`logrus`、`klog`、`zerolog` 等提供的轻量级异步日志库。它设计简单，易于使用，提供了一系列便利的功能，帮助你快速设置日志。
>
> 你可以使用命令 `go get github.com/shengyanli1982/law` 来安装 `law`。

**示例**

```go
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/law"
	"github.com/shengyanli1982/orbit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})
}

func main() {
	// 使用 os.Stdout 和配置创建一个新的 WriteAsyncer 实例
	// Create a new WriteAsyncer instance using os.Stdout and the configuration
	w := law.NewWriteAsyncer(os.Stdout, nil)

	// 使用 defer 语句确保在 main 函数退出时停止 WriteAsyncer
	// Use a defer statement to ensure that WriteAsyncer is stopped when the main function exits
	defer w.Stop()

	// 创建一个 zapcore.EncoderConfig 实例，用于配置 zap 的编码器
	// Create a zapcore.EncoderConfig instance to configure the encoder of zap
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",                         // 消息的键名
		LevelKey:       "level",                       // 级别的键名
		NameKey:        "logger",                      // 记录器名的键名
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 级别的编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // 时间的编码器
		EncodeDuration: zapcore.StringDurationEncoder, // 持续时间的编码器
	}

	// 使用 WriteAsyncer 创建一个 zapcore.WriteSyncer 实例
	// Create a zapcore.WriteSyncer instance using WriteAsyncer
	zapAsyncWriter := zapcore.AddSync(w)

	// 使用编码器配置和 WriteSyncer 创建一个 zapcore.Core 实例
	// Create a zapcore.Core instance using the encoder configuration and WriteSyncer
	zapCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapAsyncWriter, zapcore.DebugLevel)

	// 使用 Core 创建一个 zap.Logger 实例
	// Create a zap.Logger instance using Core
	zapLogger := zap.New(zapCore)

	// 创建一个新的 Orbit 配置，并设置访问日志事件函数
	// Create a new Orbit configuration and set the access log event function
	config := orbit.NewConfig().WithLogger(zapLogger)

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
$ go run demo.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] GET    /demo                     --> main.(*service).RegisterGroup.func1 (5 handlers)
{"level":"info","msg":"http server is ready","address":"127.0.0.1:8080"}
{"level":"info","msg":"http server access log","id":"","ip":"127.0.0.1","endpoint":"127.0.0.1:50940","path":"/demo","method":"GET","code":200,"status":"OK","latency":"20.445µs","agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36","query":"","reqContentType":"","reqBody":""}
{"level":"info","msg":"http server is shutdown","address":"127.0.0.1:8080"}
```

## 9. Prometheus 指标

`orbit` 支持 `prometheus` 指标。您可以使用 `EnableMetric` 来启用它。以下是使用 `demo` 服务收集 `demo` 指标的示例。

> [!TIP]
>
> 使用 curl http://127.0.0.1:8080/metrics 获取指标。

**示例**

```go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})
}

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项，并启用 metric
	// Create a new Orbit feature options and enable metric
	opts := orbit.NewOptions().EnableMetric()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
# HELP orbit_http_request_latency_seconds HTTP request latencies in seconds.
# TYPE orbit_http_request_latency_seconds gauge
orbit_http_request_latency_seconds{method="GET",path="/demo",status="200"} 0
# HELP orbit_http_request_latency_seconds_histogram HTTP request latencies in seconds(Histogram).
# TYPE orbit_http_request_latency_seconds_histogram histogram
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="0.1"} 3
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="0.5"} 3
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="1"} 3
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="2"} 3
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="5"} 3
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="10"} 3
orbit_http_request_latency_seconds_histogram_bucket{method="GET",path="/demo",status="200",le="+Inf"} 3
orbit_http_request_latency_seconds_histogram_sum{method="GET",path="/demo",status="200"} 0
orbit_http_request_latency_seconds_histogram_count{method="GET",path="/demo",status="200"} 3
```

## 10. 重复读取请求/响应体

`orbit` 支持重复读取请求/响应体。默认情况下，此行为已启用，无需额外配置。以下是如何使用 `demo` 服务来重复读取请求/响应体的示例。

### 10.1 重复读取请求体

您可以使用 `httptool.GenerateRequestBody` 方法获取请求体的字节并进行缓存。这样，您可以在需要时读取缓存的字节。

> [!IMPORTANT]
>
> 请求体是一个 `io.ReadCloser`，它是一个只能读取一次的流。如果您需要再次读取它，请不要直接读取它，而是使用 `orbit` 进行缓存。

**示例**

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 POST 方法的处理函数
	// Register a POST method handler function on the "/demo" path
	g.POST("/demo", func(c *gin.Context) {
		// 重复读取请求体内容 20 次
		// Repeat the read request body content 20 times
		for i := 0; i < 20; i++ {
			// 生成请求体
			// Generate the request body
			if body, err := httptool.GenerateRequestBody(c); err != nil {
				// 如果生成请求体出错，返回 HTTP 状态码 500 和错误信息
				// If there is an error generating the request body, return HTTP status code 500 and the error message
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				// 如果生成请求体成功，返回 HTTP 状态码 200 和请求体内容
				// If the request body is successfully generated, return HTTP status code 200 and the request body content
				c.String(http.StatusOK, fmt.Sprintf(">> %d, %s\n", i, string(body)))
			}
		}
	})
}

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 模拟一个请求
	// Simulate a request
	resp, _ := http.Post("http://localhost:8080/demo", "text/plain", io.Reader(bytes.NewBuffer([]byte("demo"))))
	defer resp.Body.Close()

	// 打印响应体
	// Print the response body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
$ go run demo.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] POST   /demo                     --> main.(*service).RegisterGroup.func1 (5 handlers)
{"level":"INFO","time":"2024-01-13T10:27:37.531+0800","logger":"default","caller":"orbit/gin.go:165","message":"http server is ready","address":"127.0.0.1:8080"}
{"level":"INFO","time":"2024-01-13T10:27:37.534+0800","logger":"default","caller":"log/default.go:10","message":"http server access log","id":"","ip":"127.0.0.1","endpoint":"127.0.0.1:58618","path":"/demo","method":"POST","code":200,"status":"OK","latency":"50.508µs","agent":"Go-http-client/1.1","query":"","reqContentType":"text/plain","reqBody":""}

>> 0, demo
>> 1, demo
>> 2, demo
>> 3, demo
>> 4, demo
>> 5, demo
>> 6, demo
>> 7, demo
>> 8, demo
>> 9, demo
>> 10, demo
>> 11, demo
>> 12, demo
>> 13, demo
>> 14, demo
>> 15, demo
>> 16, demo
>> 17, demo
>> 18, demo
>> 19, demo

{"level":"INFO","time":"2024-01-13T10:28:07.537+0800","logger":"default","caller":"orbit/gin.go:190","message":"http server is shutdown","address":"127.0.0.1:8080"}
```

### 10.2 重复读取响应体

`httptool.GenerateResponseBody` 方法可用于从缓存中获取响应体的字节。需要注意的是，在写入实际的响应体之后，如使用 `c.String(http.StatusOK, "demo")`，才能调用 `httptool.GenerateResponseBody`。

> [!NOTE]
>
> 响应体总是被写入到 `io.Writer`，因此无法直接读取。如果需要读取它，可以使用 `orbit` 进行缓存。
>
> `httptool.GenerateResponseBody` 经常在自定义中间件中使用，以获取响应体的字节并执行其他操作。

**示例**

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

// customMiddleware 函数定义了一个自定义的中间件
// The customMiddleware function defines a custom middleware
func customMiddleware() gin.HandlerFunc {
	// 返回一个 Gin 的 HandlerFunc
	// Return a Gin HandlerFunc
	return func(c *gin.Context) {
		// 调用下一个中间件或处理函数
		// Call the next middleware or handler function
		c.Next()

		// 从上下文中获取响应体缓冲区
		// Get the response body buffer from the context
		for i := 0; i < 20; i++ {
			// 生成响应体
			// Generate the response body
			body, _ := httptool.GenerateResponseBody(c)
			// 打印响应体
			// Print the response body
			fmt.Printf("# %d, %s\n", i, string(body))
		}
	}
}

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})
}

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项，并启用 metric
	// Create a new Orbit feature options and enable metric
	opts := orbit.NewOptions().EnableMetric()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的中间件
	// Register a custom middleware
	engine.RegisterMiddleware(customMiddleware())

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 模拟一个请求
	// Simulate a request
	resp, _ := http.Get("http://localhost:8080/demo")
	defer resp.Body.Close()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
```

**执行结果**

```bash
$ go run demo.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] GET    /metrics                  --> github.com/shengyanli1982/orbit/utils/wrapper.WrapHandlerToGin.func1 (2 handlers)
[GIN-debug] GET    /demo                     --> main.(*service).RegisterGroup.func1 (7 handlers)
{"level":"INFO","time":"2024-01-13T10:32:25.191+0800","logger":"default","caller":"orbit/gin.go:165","message":"http server is ready","address":"127.0.0.1:8080"}
{"level":"INFO","time":"2024-01-13T10:32:25.194+0800","logger":"default","caller":"log/default.go:10","message":"http server access log","id":"","ip":"127.0.0.1","endpoint":"127.0.0.1:59139","path":"/demo","method":"GET","code":200,"status":"OK","latency":"20.326µs","agent":"Go-http-client/1.1","query":"","reqContentType":"","reqBody":""}

# 0, demo
# 1, demo
# 2, demo
# 3, demo
# 4, demo
# 5, demo
# 6, demo
# 7, demo
# 8, demo
# 9, demo
# 10, demo
# 11, demo
# 12, demo
# 13, demo
# 14, demo
# 15, demo
# 16, demo
# 17, demo
# 18, demo
# 19, demo

{"level":"INFO","time":"2024-01-13T10:32:55.195+0800","logger":"default","caller":"orbit/gin.go:190","message":"http server is shutdown","address":"127.0.0.1:8080"}
```

## 11. 优雅关闭

`orbit` 引擎具备优雅关闭的功能。虽然可以使用 `engine.Stop` 方法来停止引擎，但它并不会立即停止。为了更优雅的关闭，你可以使用 [`GS`](https://github.com/shengyanli1982/gs) 项目。

> [!TIP]
>
> [`GS`](https://github.com/shengyanli1982/gs) 是一个轻量级的 Go 库，用于帮助实现优雅的关闭。它设计简单易用，提供了一系列方便的功能，帮助你快速实现优雅的关闭。
>
> 你可以使用命令 `go get github.com/shengyanli1982/gs` 来安装 `GS`。

**示例**

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/gs"
	"github.com/shengyanli1982/orbit"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})
}

func main() {
	// 创建一个新的 TerminateSignal 实例
	// Create a new TerminateSignal instance
	sig := gs.NewTerminateSignal()

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(nil, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 注册需要在终止信号发生时执行的处理函数
	// Register the handle functions to be executed when the termination signal occurs
	sig.RegisterCancelHandles(engine.Stop)

	// 等待所有的异步关闭信号
	// Wait for all asynchronous shutdown signals
	gs.WaitForAsync(sig)
}
```

**执行结果**

```bash
$ go run demo.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] GET    /demo                     --> main.(*service).RegisterGroup.func1 (5 handlers)
{"level":"INFO","time":"2024-06-24T17:27:04.646+0800","logger":"default","caller":"orbit/gin.go:334","message":"http server is ready","address":"127.0.0.1:8080"}
{"level":"INFO","time":"2024-06-24T17:27:06.068+0800","logger":"default","caller":"log/default.go:11","message":"http server access log","id":"","ip":"127.0.0.1","endpoint":"127.0.0.1:51540","path":"/demo","method":"GET","code":200,"status":"OK","latency":"25.344µs","agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36","query":"","reqContentType":"","reqBody":""}
{"level":"INFO","time":"2024-06-24T17:27:12.196+0800","logger":"default","caller":"orbit/gin.go:373","message":"http server is shutdown","address":"127.0.0.1:8080"}
```
