package metric

import "github.com/go-logr/logr"

// ErrorLog 结构体包装了 logr.Logger，用于处理 Prometheus 指标错误日志。
// The ErrorLog struct wraps logr.Logger to handle Prometheus metrics error logging.
type ErrorLog struct {
	l *logr.Logger // logger 实例 (logger instance)
}

// NewErrorLog 函数返回一个新的 ErrorLog 实例。
// The NewErrorLog function returns a new ErrorLog instance.
func NewErrorLog(l *logr.Logger) *ErrorLog {
	return &ErrorLog{l: l}
}

// Println 方法实现了 Prometheus 客户端所需的错误日志接口。
// The Println method implements the error logging interface required by Prometheus client.
func (e *ErrorLog) Println(v ...interface{}) {
	// 使用 Info 级别记录 Prometheus 指标错误
	// Log Prometheus metric errors at Info level
	(*e.l).Info("prometheus metric error", v)
}
