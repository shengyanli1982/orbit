package common

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
)

// RequestBodyBufferPool is a buffer pool for request bodies.
var RequestBodyBufferPool = bp.NewBufferPool(0)

// ResponseBodyBufferPool is a buffer pool for response bodies.
var ResponseBodyBufferPool = bp.NewBufferPool(0)

// LogEventPool is a pool for log events.
var LogEventPool = bp.NewLogEventPool()
