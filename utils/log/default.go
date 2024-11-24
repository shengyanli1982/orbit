package log

import (
	"github.com/go-logr/logr"
)

// DefaultAccessEventFunc 是默认的访问日志事件函数。
// DefaultAccessEventFunc is the default access log event function.
func DefaultAccessEventFunc(logger *logr.Logger, event *LogEvent) {
	logger.Info(
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

// DefaultRecoveryEventFunc 是默认的恢复日志事件函数。
// DefaultRecoveryEventFunc is the default recovery log event function.
func DefaultRecoveryEventFunc(logger *logr.Logger, event *LogEvent) {
	logger.Error(
		event.Error,
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
		"errorStack", event.ErrorStack,
	)
}
