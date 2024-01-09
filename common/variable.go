package common

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
)

var (
	RequestBodyBufferPool  = bp.NewBufferPool(0)
	ResponseBodyBufferPool = bp.NewBufferPool(0)
	LogEventPool           = bp.NewLogEventPool()
)
