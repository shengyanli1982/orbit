package common

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
	log "github.com/shengyanli1982/orbit/utils/log"
)

// RequestBodyBufferPool 是用于请求体的缓冲池。
// RequestBodyBufferPool is a buffer pool for request bodies.
var RequestBodyBufferPool = bp.NewBufferPool(0)

// ResponseBodyBufferPool 是用于响应体的缓冲池。
// ResponseBodyBufferPool is a buffer pool for response bodies.
var ResponseBodyBufferPool = bp.NewBufferPool(0)

// LogEventPool 是用于日志事件的池。
// LogEventPool is a pool for log events.
var LogEventPool = bp.NewLogEventPool()

// DefaultConsoleLogger 是默认的控制台日志记录器。
// DefaultConsoleLogger is the default console logger.
var DefaultConsoleLogger = log.NewLogger(nil)

// DefaultSugeredLogger 是默认的带糖的日志记录器。
// DefaultSugeredLogger is the default sugared logger.
var DefaultSugeredLogger = DefaultConsoleLogger.GetZapSugaredLogger().Named(log.DefaultLoggerName)
