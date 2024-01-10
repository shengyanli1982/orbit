<div align="center">
	<h1>orbit</h1>
	<img src="assets/logo.png" alt="logo" width="300px">
	<h4>A lightweight http web service wrapper framework.</h4>
</div>

# Introduction

`orbit` is a lightweight http web service wrapper framework. It is designed to be simple and easy to use. It is based on [`gin`](https://github.com/gin-gonic/gin), [`zap`](https://github.com/uber-go/zap) and [`prometheus`](github.com/prometheus/client_golang). It provides a series of convenient features to help you quickly build a web service.

Why is it called `orbit`? Because it is a lightweight framework, it is like a satellite orbiting the earth, and the orbit is your business logic.

Why not use `gin` directly? Because `gin` is too simple, if you want to start a web service, you need to do a lot of work, such as logging, monitoring, etc. `orbit` is based on `gin`, and provides a series of convenient features to help you quickly build a web service.

# Advantages

-   Lightweight, easy to use, easy to learn
-   Support `zap` logging, `async` and `sync` logging
-   Support `prometheus` monitoring
-   Support `swagger` api document
-   Support `graceful` shutdown
-   Support `cors` middleware
-   Support auto recover from panic
-   Support custom middleware
-   Support custom router group
-   Support custom define access log format and fields

# Installation

```bash
go get github.com/shengyanli1982/orbit
```

# Quick Start

`orbit` is very easy to use, you can quickly build a web service in a few minutes. usually, you only need to do the following:

1. Create `orbit` start configuration
2. Create `orbit` feature options
3. Create `orbit` instance

**Default URL PATH**

> [!NOTE]
> Default url path is system default, you can not change it and it is not in your control.

-   `/metrics` - prometheus metrics
-   `/swagger/*any` - swagger api document
-   `/debug/pprof/*any` - pprof debug
-   `/ping` - health check

## 1. Config

The `orbit` has some config options, you can set it before start `orbit` instance.

-   `WithSugaredLogger` - use `zap` sugared logger, default is `DefaultSugeredLogger`
-   `WithLogger` - use `zap` logger, default is `DefaultConsoleLogger`
-   `WithAddress` - http server listen address, default is `127.0.0.0`
-   `WithPort` - http server listen port, default is `8080`
-   `WithRelease` - http server release mode, default is `false`
-   `WithHttpReadTimeout` - http server read timeout, default is `15s`
-   `WithHttpWriteTimeout` - http server write timeout, default is `15s`
-   `WithHttpReadHeaderTimeout` - http server read header timeout, default is `15s`
-   `WithAccessLogEventFunc` - http server access log event func, default is `DefaultAccessEventFunc`
-   `WithRecoveryLogEventFunc` - http server recovery log event func, default is `DefaultRecoveryEventFunc`
-   `WithPrometheusRegistry` - http server prometheus registry, default is `prometheus.DefaultRegister`

You can use `NewConfig` to create a default config, and use `WithXXX` to set config options. `DefaultConfig` is alias of `NewConfig()`.

## 2. Feature

Also `orbit` has some feature options, you can set it before start `orbit` instance.

-   `EnablePProf` - enable pprof debug, default is `false`
-   `EnableSwagger` - enable swagger api document, default is `false`
-   `EnableMetric` - enable prometheus metrics, default is `false`
-   `EnableRedirectTrailingSlash` - enable redirect trailing slash, default is `false`
-   `EnableRedirectFixedPath` - enable redirect fixed path, default is `false`
-   `EnableForwardedByClientIp` - enable forwarded by client ip, default is `false`
-   `EnableRecordRequestBody` - enable record request body, default is `false`

You can use `NewOptions` to create a null feature, and use `EnableXXX` to set feature options.

-   `DebugOptions` use for debug, it is alias of `NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()`.
-   `ReleaseOptions` use for release, it is alias of `NewOptions().EnableMetric()`.

> [!NOTE]
> Here is a best recommendation, you can use `DebugOptions` for debug, and use `ReleaseOptions` for release.

## 3. Instance

After you create `orbit` config and feature options, you can create `orbit` instance.

> [!IMPORTANT]
> When you `Run` the `orbit` instance, it will not block the current goroutine which mean you can do other things after `Run` the `orbit` instance.
>
> If you want to block the current goroutine, you can use project [`GS`](https://github.com/shengyanli1982/gs) to provide a `Waitting` to block the current goroutine.

> [!TIP]
> Here is a way to lazy. You can use `NewHttpService` to wrap `func(*gin.RouterGroup)` to `Service` interface implementation.
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
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
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

Because `orbit` is based on `gin`, so you can use `gin` middleware directly. So you can use custom middleware to do some custom things. For example, you can use `cors` middleware to support `cors` request.

Here is a example to use `demo` middleware to print `>>>>>>!!! demo` in the console.

**Example**

```go
package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
)

func customMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println(">>>>>>!!! demo")
		c.Next()
	}
}

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/demo", func(c *gin.Context) {})
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions().EnableMetric()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom middleware.
	engine.RegisterMiddleware(customMiddleware())

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
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

Custom router group is a very useful feature, you can use it to register a custom router group. You can use it to register a custom router group for `demo` service.

**eg:** `/demo` and `/demo/test`

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

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// Register a custom router group.
	g = g.Group("/demo")

	// /demo
	g.GET(ocom.EmptyURLPath, func(c *gin.Context) {
		c.String(http.StatusOK, "demo")
	})

	// /demo/test
	g.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions().EnableMetric()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
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

Http server access log is very important, you can use `orbit` to custom access log format and fields. Here is a example to custom access log format and fields.

**Default LogEvent Fields**

```go
// LogEvent represents a log event.
type LogEvent struct {
	Message        string `json:"message,omitempty" yaml:"message,omitempty"`               // Message contains the log message.
	ID             string `json:"id,omitempty" yaml:"id,omitempty"`                         // ID contains the unique identifier of the log event.
	IP             string `json:"ip,omitempty" yaml:"ip,omitempty"`                         // IP contains the IP address of the client.
	EndPoint       string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`             // EndPoint contains the endpoint of the request.
	Path           string `json:"path,omitempty" yaml:"path,omitempty"`                     // Path contains the path of the request.
	Method         string `json:"method,omitempty" yaml:"method,omitempty"`                 // Method contains the HTTP method of the request.
	Code           int    `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`         // Code contains the HTTP status code of the response.
	Status         string `json:"status,omitempty" yaml:"status,omitempty"`                 // Status contains the status message of the response.
	Latency        string `json:"latency,omitempty" yaml:"latency,omitempty"`               // Latency contains the request latency.
	Agent          string `json:"agent,omitempty" yaml:"agent,omitempty"`                   // Agent contains the user agent of the client.
	ReqContentType string `json:"reqContentType,omitempty" yaml:"reqContentType,omitempty"` // ReqContentType contains the content type of the request.
	ReqQuery       string `json:"query,omitempty" yaml:"query,omitempty"`                   // ReqQuery contains the query parameters of the request.
	ReqBody        string `json:"reqBody,omitempty" yaml:"reqBody,omitempty"`               // ReqBody contains the request body.
	Error          any    `json:"error,omitempty" yaml:"error,omitempty"`                   // Error contains the error object.
	ErrorStack     string `json:"errorStack,omitempty" yaml:"errorStack,omitempty"`         // ErrorStack contains the stack trace of the error.
}

```

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

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// /demo
	g.GET("/demo", func(c *gin.Context) {
		c.String(http.StatusOK, "demo")
	})
}

func customAccessLogger(logger *zap.SugaredLogger, event *log.LogEvent) {
	logger.Infow("access log", "path", event.Path, "method", event.Method)
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig().WithAccessLogEventFunc(customAccessLogger)

	// Create a new orbit feature options.
	opts := orbit.NewOptions()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
	engine.Stop()
}
```

**Result**

```bash
{"level":"INFO","time":"2024-01-10T20:22:01.244+0800","logger":"default","caller":"accesslog/demo.go:24","message":"access log","path":"/demo","method":"GET"}
```

## 7. Custom Recovery Log

Http server recovery log give you a chance to know what happened when your service panic. You can use `orbit` to custom recovery log format and fields. Here is a example to custom recovery log format and fields.

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

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// /demo
	g.GET("/demo", func(c *gin.Context) {
		panic("demo")
	})
}

func customRecoveryLogger(logger *zap.SugaredLogger, event *log.LogEvent) {
	logger.Infow("recovery log", "path", event.Path, "method", event.Method, "error", event.Error, "errorStack", event.ErrorStack)
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig().WithRecoveryLogEventFunc(customRecoveryLogger)

	// Create a new orbit feature options.
	opts := orbit.NewOptions()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
	engine.Stop()
}

```

**Result**

```bash
{"level":"INFO","time":"2024-01-10T20:27:10.041+0800","logger":"default","caller":"recoverylog/demo.go:22","message":"recovery log","path":"/demo","method":"GET","error":"demo","errorStack":"goroutine 6 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x65\ngithub.com/shengyanli1982/orbit/internal/middleware.Recovery.func1.1()\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:145 +0x559\npanic({0x170ec80, 0x191cb70})\n\t/usr/local/go/src/runtime/panic.go:884 +0x213\nmain.(*service).RegisterGroup.func1(0x0?)\n\t/Volumes/DATA/programs/GolandProjects/orbit/example/recoverylog/demo.go:17 +0x27\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.AccessLogger.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:59 +0x1a5\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.Cors.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:35 +0x139\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.BodyBuffer.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/buffer.go:18 +0x92\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/shengyanli1982/orbit/internal/middleware.Recovery.func1(0xc0001e6300)\n\t/Volumes/DATA/programs/GolandProjects/orbit/internal/middleware/system.go:166 +0x82\ngithub.com/gin-gonic/gin.(*Context).Next(...)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/context.go:173\ngithub.com/gin-gonic/gin.(*Engine).handleHTTPRequest(0xc0000076c0, 0xc0001e6300)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/gin.go:616 +0x66b\ngithub.com/gin-gonic/gin.(*Engine).ServeHTTP(0xc0000076c0, {0x1924a30?, 0xc0000c02a0}, 0xc0001e6200)\n\t/Volumes/CACHE/programs/gopkgs/pkg/mod/github.com/gin-gonic/gin@v1.8.2/gin.go:572 +0x1dd\nnet/http.serverHandler.ServeHTTP({0xc00008be30?}, {0x1924a30, 0xc0000c02a0}, 0xc0001e6200)\n\t/usr/local/go/src/net/http/server.go:2936 +0x316\nnet/http.(*conn).serve(0xc0000962d0, {0x19253e0, 0xc00008bd40})\n\t/usr/local/go/src/net/http/server.go:1995 +0x612\ncreated by net/http.(*Server).Serve\n\t/usr/local/go/src/net/http/server.go:3089 +0x5ed\n"}
```

## 8. Prometheus Metrics

`orbit` support `prometheus` metrics, you can use `EnableMetric` to enable it. Here is a example to use `demo` service to collect `demo` metrics.

> [!TIP]
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

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// /demo
	g.GET("/demo", func(c *gin.Context) {
		c.String(http.StatusOK, "demo")
	})

}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions().EnableMetric()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
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
