package metric

import "github.com/go-logr/logr"

// 包装了 logr.Logger，用于处理 Prometheus 指标错误日志
type ErrorLog struct {
	l *logr.Logger // logger 实例
}

// 返回一个新的 ErrorLog 实例
func NewErrorLog(l *logr.Logger) *ErrorLog {
	return &ErrorLog{l: l}
}

// 实现了 Prometheus 客户端所需的错误日志接口
func (e *ErrorLog) Println(v ...interface{}) {
	(*e.l).Info("prometheus metric error", v)
}
