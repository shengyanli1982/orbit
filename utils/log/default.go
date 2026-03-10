package log

import (
	"github.com/go-logr/logr"
)

// 默认的访问日志事件函数
func DefaultAccessEventFunc(logger *logr.Logger, event *LogEvent) {
	logger.Info(
		event.Message,
		"id", event.ID, // backward compatibility
		"ip", event.IP, // backward compatibility
		"request_id", event.ID,
		"client_ip", event.IP,
		"forwarded_for", event.ForwardedFor,
		"endpoint", event.EndPoint,
		"path", event.Path,
		"method", event.Method,
		"code", event.Code,
		"status", event.Status,
		"latency", event.Latency,
		"latency_ms", event.LatencyMs,
		"agent", event.Agent,
		"query", event.ReqQuery,
		"reqContentType", event.ReqContentType,
		"reqBody", event.ReqBody,
	)
}

// 默认的恢复日志事件函数
func DefaultRecoveryEventFunc(logger *logr.Logger, event *LogEvent) {
	logger.Error(
		event.Error,
		event.Message,
		"id", event.ID, // backward compatibility
		"ip", event.IP, // backward compatibility
		"request_id", event.ID,
		"client_ip", event.IP,
		"forwarded_for", event.ForwardedFor,
		"endpoint", event.EndPoint,
		"path", event.Path,
		"method", event.Method,
		"code", event.Code,
		"status", event.Status,
		"latency", event.Latency,
		"latency_ms", event.LatencyMs,
		"agent", event.Agent,
		"query", event.ReqQuery,
		"reqContentType", event.ReqContentType,
		"reqBody", event.ReqBody,
		"errorStack", event.ErrorStack,
	)
}
