package metric

import "github.com/go-logr/logr"

type ErrorLog struct {
	l *logr.Logger
}

func NewErrorLog(l *logr.Logger) *ErrorLog {
	return &ErrorLog{l: l}
}

func (e *ErrorLog) Println(v ...interface{}) {
	(*e.l).Info("prometheus metric error", v)
}
