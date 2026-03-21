<div align="center">
  <img src="assets/logo.png" alt="orbit logo" width="500px">

# Orbit

**Build production-ready HTTP services on Gin, faster.**

Orbit gives Go teams a clean service scaffold with practical defaults for observability, reliability, and operations.

</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/orbit)](https://goreportcard.com/report/github.com/shengyanli1982/orbit)
[![Build Status](https://github.com/shengyanli1982/orbit/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/orbit/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/orbit.svg)](https://pkg.go.dev/github.com/shengyanli1982/orbit)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shengyanli1982/orbit)

## Why Orbit

When plain Gin feels too bare for real services, Orbit fills the gap without getting in your way.

- **Ship faster**: health endpoint, metrics, recovery, and access logging are built in.
- **Stay familiar**: keep Gin handlers, router groups, and middleware style.
- **Operate with confidence**: graceful shutdown, timeout controls, proxy/CORS config, and strong observability defaults.
- **Scale cleanly**: split responsibilities with `Config` (server behavior) + `Options` (feature flags) + `Service` (route modules).

## What You Get Out of the Box

| Area               | What Orbit Provides                                                           |
| ------------------ | ----------------------------------------------------------------------------- |
| Runtime            | `Engine` lifecycle (`Run`, `Stop`), service and middleware registration       |
| Built-in endpoints | `/ping`, optional `/metrics`, `/docs/*any`, `/debug/pprof/*any`               |
| Observability      | `logr`-based logging, Prometheus middleware, panic recovery, access logging   |
| Reliability        | graceful shutdown, read/write/header/idle timeout controls, max header limits |
| Extensibility      | `Service` interface + `NewHttpService` helper + custom middleware hooks       |
| JSON codec         | selectable `encoding/json`, `jsoniter`, `sonic` via build tags                |

## Quick Start (Copy-Paste)

```bash
go get github.com/shengyanli1982/orbit
```

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
)

func main() {
	cfg := orbit.NewConfig()
	opts := orbit.ReleaseOptions() // metric enabled, health check enabled

	engine := orbit.NewEngine(cfg, opts)
	engine.RegisterService(orbit.NewHttpService(func(g *gin.RouterGroup) {
		g.GET("/demo", func(c *gin.Context) {
			c.String(http.StatusOK, "demo")
		})
	}))
	defer engine.Stop()

	engine.Run() // non-blocking
	select {}
}
```

Test it:

```bash
curl -i http://127.0.0.1:8080/ping
```

## Architecture Snapshot

- **`Config`**: address, port, timeouts, CORS, trusted proxies, logger, Prometheus registry.
- **`Options`**: switches like `EnableMetric`, `EnableSwagger`, `EnablePProf`, `EnableRecordRequestBody`.
- **`Engine`**: wires middleware/services and owns lifecycle.
- **`Service`**: feature modules register routes through `RegisterGroup(*gin.RouterGroup)`.

Request pipeline (high-level):

1. base middleware (`Recovery` -> `BodyBuffer` -> `CorsWithPolicy`)
2. metrics middleware (when enabled)
3. custom middleware (`RegisterMiddleware`)
4. access logger
5. user handlers (`RegisterService`)

## Built-in Endpoints

| Endpoint            | Default | Toggle                          |
| ------------------- | ------- | ------------------------------- |
| `/ping`             | on      | on by default in `NewOptions()` |
| `/metrics`          | off     | `EnableMetric()`                |
| `/docs/*any`        | off     | `EnableSwagger()`               |
| `/debug/pprof/*any` | off     | `EnablePProf()`                 |

## JSON Backend Selection

Orbit selects JSON backend through build tags (`internal/codec/json`):

| Backend                       | Build Tag  | Notes                    |
| ----------------------------- | ---------- | ------------------------ |
| `encoding/json`               | (none)     | default                  |
| `github.com/json-iterator/go` | `jsoniter` | compatible alternative   |
| `github.com/bytedance/sonic`  | `sonic`    | high-performance backend |

Selection precedence:

- no tag => `encoding/json`
- `-tags jsoniter` => `jsoniter`
- `-tags sonic` => `sonic`
- `-tags "jsoniter sonic"` => `sonic` wins

```bash
go test ./...
go test -tags jsoniter ./...
go test -tags sonic ./...
```

## Performance & Reliability Notes

- `sync.Pool`-based buffer pools for request/response body buffering and log event reuse.
- Path-normalized metric labels (`c.FullPath()`) to reduce cardinality risk.
- Full timeout and header-limit controls for predictable resource behavior.
- Graceful shutdown sequence designed for in-flight request safety.

## Examples

- [`examples/simpleserver`](./examples/simpleserver)
- [`examples/middleware`](./examples/middleware)
- [`examples/metric`](./examples/metric)
- [`examples/routegroup`](./examples/routegroup)
- [`examples/repeat/request`](./examples/repeat/request)
- [`examples/repeat/response`](./examples/repeat/response)
- [`examples/gracefulstop`](./examples/gracefulstop)

## Learn More

- DeepWiki (full guides and architecture): <https://deepwiki.com/shengyanli1982/orbit>
- Go API reference: <https://pkg.go.dev/github.com/shengyanli1982/orbit>

## License

[MIT](./LICENSE)
