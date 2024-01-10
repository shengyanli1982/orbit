package log

import (
	"github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap"
)

// DefaultAccessEventFunc is the default access log event function.
func DefaultAccessEventFunc(logger *zap.SugaredLogger, event *log.LogEvent) {
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
func DefaultRecoveryEventFunc(logger *zap.SugaredLogger, event *log.LogEvent) {
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
