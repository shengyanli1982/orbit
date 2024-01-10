package common

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
	log "github.com/shengyanli1982/orbit/utils/log"
)

// RequestBodyBufferPool is a buffer pool for request bodies.
var RequestBodyBufferPool = bp.NewBufferPool(0)

// ResponseBodyBufferPool is a buffer pool for response bodies.
var ResponseBodyBufferPool = bp.NewBufferPool(0)

// LogEventPool is a pool for log events.
var LogEventPool = bp.NewLogEventPool()

// DefaultConsoleLogger is the default console logger.
var DefaultConsoleLogger = log.NewLogger(nil)

// DefaultSugeredLogger is the default sugared logger.
var DefaultSugeredLogger = DefaultConsoleLogger.GetZapSugaredLogger().Named(log.DefaultLoggerName)
