package log

import (
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"go.uber.org/zap"
)

// DefaultAccessEventFunc is the default access log event function.
func DefaultAccessEventFunc(logger *zap.SugaredLogger, event *bp.LogEvent) {
	logger.Infow(
		event.Message,
		"id", event.ID,
		"ip", event.IP,
		"endpoint", event.EndPoint,
		"path", event.Path,
		"method", event.Method,
		"code", event.Code,
		"status", event.Status,
		"latency", event.Latency,
		"agent", event.Agent,
		"query", event.ReqQuery,
		"reqContentType", event.ReqContentType,
		"reqBody", event.ReqBody,
	)
}

// DefaultRecoveryEventFunc is the default recovery log event function.
func DefaultRecoveryEventFunc(logger *zap.SugaredLogger, event *bp.LogEvent) {
	logger.Errorw(
		event.Message,
		"id", event.ID,
		"ip", event.IP,
		"endpoint", event.EndPoint,
		"path", event.Path,
		"method", event.Method,
		"code", event.Code,
		"status", event.Status,
		"latency", event.Latency,
		"agent", event.Agent,
		"query", event.ReqQuery,
		"reqContentType", event.ReqContentType,
		"reqBody", event.ReqBody,
		"error", event.Error,
		"errorStack", event.ErrorStack,
	)
}
