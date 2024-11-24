English | [中文](./README_CN.md)

<div align="center">
	<img src="assets/logo.png" alt="logo" width="500px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/orbit)](https://goreportcard.com/report/github.com/shengyanli1982/orbit)
[![Build Status](https://github.com/shengyanli1982/orbit/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/orbit/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/orbit.svg)](https://pkg.go.dev/github.com/shengyanli1982/orbit)

# Introduction

`orbit` is a lightweight HTTP web service wrapper framework designed for simplicity and ease of use. It provides a suite of convenient features to help you quickly build and maintain a web service.

The name `orbit` reflects the framework's goal of encapsulating the complexities of building a web service, allowing you to focus on your core business logic, much like a satellite smoothly orbiting the Earth.

### Why Not Use `gin` Directly?

While `gin` is an excellent framework, it requires additional setup for logging and monitoring. `orbit` is built on top of `gin`, providing these features out-of-the-box, streamlining the process of starting a web service.

# Advantages

-   Lightweight and user-friendly
-   Supports `zap` and `klog` logging with both `async` and `sync` modes
-   Integrates `prometheus` for monitoring
-   Includes `swagger` API documentation support
-   Graceful server shutdown
-   `cors` middleware for cross-origin requests
-   Automatic panic recovery
-   Customizable middleware
-   Flexible router groups
-   Customizable access log format and fields
-   Supports repeat reading of request/response body and caching

# Installation

```bash
go get github.com/shengyanli1982/orbit
```

# Quick Start

`orbit` is designed for quick and easy web service development. Follow these simple steps:

1. Create the `orbit` configuration.
2. Define the `orbit` feature options.
3. Create an `orbit` instance.

**Default URL Paths**

> [!NOTE]
>
> The default URL paths are system-defined and cannot be changed.

-   `/metrics` - Prometheus metrics
-   `/swagger/*any` - Swagger API documentation
-   `/debug/pprof/*any` - PProf debug
-   `/ping` - Health check

## 1. Configuration

`orbit` provides several configuration options that can be set before starting the `orbit` instance.

-   `WithLogger` - Use `logr` logger (default: `DefaultConsoleLogger`).
-   `WithAddress` - HTTP server listen address (default: `127.0.0.1`).
-   `WithPort` - HTTP server listen port (default: `8080`).
-   `WithRelease` - HTTP server release mode (default: `false`).
-   `WithHttpReadTimeout` - HTTP server read timeout (default: `15s`).
-   `WithHttpWriteTimeout` - HTTP server write timeout (default: `15s`).
-   `WithHttpReadHeaderTimeout` - HTTP server read header timeout (default: `15s`).
-   `WithAccessLogEventFunc` - HTTP server access log event function (default: `DefaultAccessEventFunc`).
-   `WithRecoveryLogEventFunc` - HTTP server recovery log event function (default: `DefaultRecoveryEventFunc`).
-   `WithPrometheusRegistry` - HTTP server Prometheus registry (default: `prometheus.DefaultRegister`).

You can use `NewConfig` to create a default configuration and `WithXXX` methods to set the configuration options. `DefaultConfig` is an alias for `NewConfig()`.

> [!IMPORTANT]
>
> The server has a default shutdown timeout of 10 seconds. During shutdown, it will:
>
> 1. Stop accepting new requests
> 2. Wait for ongoing requests to complete
> 3. Close all active connections
> 4. Unregister metrics collectors (if enabled)

## 2. Features

`orbit` provides several feature options that can be set before starting the `orbit` instance:

-   `EnablePProf` - enable pprof debug (default: `false`)
-   `EnableSwagger` - enable swagger API documentation (default: `false`)
-   `EnableMetric` - enable Prometheus metrics (default: `false`)
-   `EnableRedirectTrailingSlash` - enable redirect trailing slash (default: `false`)
-   `EnableRedirectFixedPath` - enable redirect fixed path (default: `false`)
-   `EnableForwardedByClientIp` - enable forwarded by client IP (default: `false`)
-   `EnableRecordRequestBody` - enable record request body (default: `false`)

You can use `NewOptions` to create a null feature, and use `EnableXXX` methods to set the feature options.

-   `DebugOptions` is used for debugging and is an alias of `NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()`.
-   `ReleaseOptions` is used for release and is an alias of `NewOptions().EnableMetric()`.

> [!NOTE]
>
> It is recommended to use `DebugOptions` for debugging and `ReleaseOptions` for release.

## 3. Creating an Instance

Once you have created the `orbit` configuration and feature options, you can create an `orbit` instance.

> [!IMPORTANT]
>
> When you run the `orbit` instance, it will not block the current goroutine. This means you can continue doing other things after running the `orbit` instance.
>
> If you want to block the current goroutine, you can use the project [`GS`](https://github.com/shengyanli1982/gs) to provide a `Waitting` function to block the current goroutine.

> [!TIP]
>
> To simplify the process, you can use `NewHttpService` to wrap the `func(*gin.RouterGroup)` into an implementation of the `Service` interface.
>
> ```go
> NewHttpService(func(g *gin.RouterGroup) {
>   g.GET("/demo", func(c *gin.Context) {
>       c.String(http.StatusOK, "demo")
>   })
> })
> ```

**Example**

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

**Result**

```bash
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
{"level":"INFO","time":"2024-01-10T17:00:13.139+0800","logger":"default","caller":"orbit/gin.go:160","message":"http server is ready","address":"127.0.0.1:8080"}
```

**Testing**

```bash
$ curl -i http://127.0.0.1:8080/ping
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 10 Jan 2024 09:07:26 GMT
Content-Length: 7

successs
```

## 4. Custom Middleware

`orbit` is based on `gin`, so you can directly use `gin` middleware. This allows you to implement custom middleware for specific tasks. For example, you can use the `demo` middleware to print `>>>>>>!!! demo` in the console.

Here is an example of using custom middleware in `orbit`:

**Example**

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

**Result**

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

## 5. Custom Router Group

The custom router group feature in `orbit` allows you to register a custom router group for the `demo` service. For example, you can register routes like `/demo` and `/demo/test`.

**Example**

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

**Result**

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

## 6. Custom Access Log

To customize the access log format and fields in `orbit`, you can use the following example:

**Default LogEvent Fields**

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

### Log Event Structure

Each log event contains the following information:

-   `message` - Log message
-   `id` - Request ID
-   `ip` - Client IP
-   `endpoint` - Request endpoint
-   `path` - Request path
-   `method` - HTTP method
-   `code` - HTTP status code
-   `status` - HTTP status text
-   `latency` - Request latency
-   `agent` - User agent
-   `query` - Request query parameters
-   `reqContentType` - Request content type
-   `reqBody` - Request body (if enabled)

### Logging Options

-   `zap` and `klog` logger with both async and sync modes
-   Standard Go logger compatibility
-   Custom log event handlers via `WithAccessLogEventFunc`

**Example**

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

**Result**

```bash
{"level":"INFO","time":"2024-01-10T20:22:01.244+0800","logger":"default","caller":"accesslog/demo.go:24","message":"access log","path":"/demo","method":"GET"}
```

## 7. Custom Recovery Log

Http server recovery log allows you to understand what happened when your service encounters a panic. With `orbit`, you can customize the recovery log format and fields. Here is an example of how to customize the recovery log format and fields.

### Log Event Structure

Each log event contains the following information:

-   `message` - Log message
-   `id` - Request ID
-   `ip` - Client IP
-   `endpoint` - Request endpoint
-   `path` - Request path
-   `method` - HTTP method
-   `code` - HTTP status code
-   `status` - HTTP status text
-   `latency` - Request latency
-   `agent` - User agent
-   `query` - Request query parameters
-   `reqContentType` - Request content type
-   `reqBody` - Request body (if enabled)
-   `error` - Error message (for recovery events)
-   `errorStack` - Error stack trace (for recovery events)

### Logging Options

-   `zap` and `klog` logger with both async and sync modes
-   Standard Go logger compatibility
-   Custom log event handlers via `WithRecoveryLogEventFunc`

**Example**

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

**Result**

```bash
{"level":"INFO","time":"2024-01-10T20:27:10.041+0800","logger":"default","caller":"recoverylog/demo.go:22","message":"recovery log","path":"/demo","method":"GET","error":"demo","errorStack":"goroutine 6 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x65\ngithub.com/shengyanli1982/orbit/internal/middleware.Recovery.func1.1()\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:145 +0x559\npanic({0x170ec80, 0x191cb70})\n\t/usr/local/go/src/runtime/panic.go:884 +0x213\nmain.(*service).RegisterGroup.func1(0x0?)\n\t/Volumes/DATA/programs/GolandProjects/orbit/example/recoverylog/demo.go:17 +0x27\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.AccessLogger.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:59 +0x1a5\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.Cors.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:35 +0x139\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.BodyBuffer.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/buffer.go:18 +0x92\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.Recovery.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:166 +0x82\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/gin-gonic/gin.(*Engine).handleHTTPRequest(0xc0000076c0, 0xc0001e6300)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/gin.go:616 +0x66b\ngithub.com/gin-gonic/gin.(*Engine).ServeHTTP(0xc0000076c0, {0x1924a30?, 0xc0000c02a0}, 0xc0001e6200)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/gin.go:572 +0x1dd\nnet/http.serverHandler.ServeHTTP({0xc00008be30?}, {0x1924a30, 0xc0000c02a0}, 0xc0001e6200)\n\t/usr/local/go/src/net/http/server.go:2936 +0x316\nnet/http.(*conn).serve(0xc0000962d0, {0x19253e0, 0xc00008bd40})\n\t/usr/local/go/src/net/http/server.go:1995 +0x612\ncreated by net/http.(*Server).Serve\n\t/usr/local/go/src/net/http/server.go:3089 +0x5ed\n"}
```

## 8. Async Logger

`orbit` leverages the `law` project to provide an `async` logger. Here's an example of how to enable the `async` logger using the `demo` service.

> [!TIP]
>
> [`law`](https://github.com/shengyanli1982/law) is a lightweight asynchronous logger for `zap`, `logrus`, `klog`, `zerolog`, and more. It's designed for simplicity and ease of use, offering a range of convenient features to help you quickly set up a logger.
>
> You can install `law` using the command `go get github.com/shengyanli1982/law`.

**Example**

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

**Result**

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

## 9. Prometheus Metrics

`orbit` supports `prometheus` metrics. You can enable it using `EnableMetric`. Here is an example of how to collect `demo` metrics using the `demo` service.

When metrics are enabled (`EnableMetric`), `orbit` automatically collects the following HTTP metrics:

-   `orbit_http_request_count` - Total number of HTTP requests made
-   `orbit_http_request_latency_seconds_histogram` - HTTP request latencies in seconds (Histogram)
-   `orbit_http_request_latency_seconds` - HTTP request latencies in seconds (Gauge)

All metrics include the following labels:

-   `method` - HTTP method
-   `path` - Request path
-   `status` - HTTP status code

> [!TIP]
>
> Use curl http://127.0.0.1:8080/metrics to get metrics.

**Example**

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

**Result**

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

## 10. Repeat Read Request/Response Body

`orbit` supports repeating the read request/response body. By default, this behavior is enabled and requires no additional configuration. Here is an example of how to use the `demo` service to repeat read the request/response body.

### 10.1 Repeat Read Request Body

You can use the `httptool.GenerateRequestBody` method to obtain the request body bytes and cache them. This allows you to read the cached bytes when needed.

> [!IMPORTANT]
>
> The request body is an `io.ReadCloser`, which is a stream that can only be read once. If you need to read it again, do not read it directly. Instead, use `orbit` to cache it.

**Example**

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

**Result**

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

### 10.2 Repeat Read Response Body

The `httptool.GenerateResponseBody` method can be used to retrieve the response body bytes from the cache. It is important to note that you should call `httptool.GenerateResponseBody` after writing the actual response body, such as using `c.String(http.StatusOK, "demo")`.

> [!NOTE]
>
> The response body is always written to an `io.Writer`, so direct reading is not possible. If you need to read it, you can use `orbit` to cache it.
>
> `httptool.GenerateResponseBody` is often used in custom middleware to retrieve the response body bytes and perform additional actions.

**Example**

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

		// 从上下文中获取响应体缓��区
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

**Result**

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

## 11. Graceful Shutdown

The `orbit` engine is equipped with a graceful shutdown feature. The `engine.Stop` method can be utilized to halt the engine, but it doesn't immediately stop it. For a more graceful shutdown, you can employ the [`GS`](https://github.com/shengyanli1982/gs) project.

> [!TIP]
>
> [`GS`](https://github.com/shengyanli1982/gs) is a lightweight library for Go that facilitates graceful shutdowns. It is designed for simplicity and ease of use, offering a range of handy features to help you implement a graceful shutdown swiftly.
>
> You can install `GS` using the command `go get github.com/shengyanli1982/gs`.

**Example**

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

**Result**

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

## 12. `Zap` and `Klog` Logger

`orbit` supports both `zap` and `klog` loggers. You can use the `zap` logger by default, or switch to the `klog` logger by setting the `Klog` field in the `Config` structure.

### 12.1 Use `Zap` Logger

`orbir` uses the `zap` logger by default. You can customize the logging behavior by creating a new `zap` logger and setting it in the configuration.

**Example**

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

func main() {
    // 创建一个新的 Zap 日志记录器
    // Create a new Zap logger
    zapLogger := log.NewZapLogger(nil)

    // 创建一个新的 Orbit 配置，并设置 Zap 日志记录器
    // Create a new Orbit configuration and set the Zap logger
    config := orbit.NewConfig().WithLogger(zapLogger.GetLogrLogger())

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

**Result**

```bash
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] GET    /demo                     --> main.(*service).RegisterGroup.func1 (5 handlers)
{"level":"INFO","time":"2024-01-13T11:15:23.191+0800","logger":"default","caller":"orbit/gin.go:165","message":"http server is ready","address":"127.0.0.1:8080"}
{"level":"INFO","time":"2024-01-13T11:15:25.194+0800","logger":"default","caller":"log/default.go:10","message":"http server access log","id":"","ip":"127.0.0.1","endpoint":"127.0.0.1:59139","path":"/demo","method":"GET","code":200,"status":"OK","latency":"20.326µs","agent":"Go-http-client/1.1","query":"","reqContentType":"","reqBody":""}
{"level":"INFO","time":"2024-01-13T11:15:53.195+0800","logger":"default","caller":"orbit/gin.go:190","message":"http server is shutdown","address":"127.0.0.1:8080"}
```

### 12.2 Use `Klog` Logger

If you want to use the `klog` logger, you can create a new `klog` logger and set it to the configuration.

**Example**

```go
package main

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/shengyanli1982/orbit"
    "github.com/shengyanli1982/orbit/utils/log"
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
    // 创建一个新的 Klog 日志记录器
    // Create a new Klog logger
    klogLogger := log.NewLogrLogger(nil)

    // 创建一个新的 Orbit 配置，并设置 Klog 日志记录器
    // Create a new Orbit configuration and set the Klog logger
    config := orbit.NewConfig().WithLogger(klogLogger.GetLogrLogger())

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

**Result**

```bash
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> github.com/shengyanli1982/orbit.healthcheckService.func1 (1 handlers)
[GIN-debug] GET    /demo                     --> main.(*service).RegisterGroup.func1 (5 handlers)
I0113 11:20:23.191458   12345 gin.go:165] http server is ready {"address": "127.0.0.1:8080"}
I0113 11:20:25.194523   12345 default.go:10] http server access log {"id": "", "ip": "127.0.0.1", "endpoint": "127.0.0.1:59139", "path": "/demo", "method": "GET", "code": 200, "status": "OK", "latency": "20.326µs", "agent": "Go-http-client/1.1", "query": "", "reqContentType": "", "reqBody": ""}
I0113 11:20:53.195721   12345 gin.go:190] http server is shutdown {"address": "127.0.0.1:8080"}
```
