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

	// Wait for 10 seconds. you can use `curl http://127.0.0.1:8080/ping` to test.
	time.Sleep(10 * time.Second)

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

Here is a example to use `demo` middleware to print `demo` in the console.

**Example**

```go

```

**Result**

```bash

```

## 5. Custom Router Group

**Example**

```go

```

**Result**

```bash

```

## 6. Custom Access Log

**Example**

```go

```

**Result**

```bash

```

## 7. Custom Recovery Log

**Example**

```go

```

**Result**

```bash

```

## 8. Prometheus Metrics

**Example**

```go

```

**Result**

```bash

```
