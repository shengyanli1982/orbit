package common

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
	log "github.com/shengyanli1982/orbit/utils/log"
)

// 用于请求体的缓冲池
var RequestBodyBufferPool = bp.NewBufferPool(0)

// 用于响应体的缓冲池
var ResponseBodyBufferPool = bp.NewBufferPool(0)

// 用于日志事件的池
var LogEventPool = bp.NewLogEventPool()

// 默认的控制台日志记录器
var DefaultConsoleLogger = log.NewZapLogger(nil, false)

// 默认的带糖的日志记录器
var DefaultSugeredLogger = DefaultConsoleLogger.GetZapSugaredLogger().Named(log.DefaultLoggerName)

// 默认的 logr 日志记录器
var DefaultLogrLogger = DefaultConsoleLogger.GetLogrLogger().WithName(log.DefaultLoggerName)
