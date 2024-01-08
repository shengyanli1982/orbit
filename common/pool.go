package common

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
)

var (
	ReqBodyBuffPool  = bp.NewBufferPool(0)
	RespBodyBuffPool = bp.NewBufferPool(0)
	LogEventPool     = bp.NewLogEventPool()
)
