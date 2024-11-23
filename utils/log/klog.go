package log

import (
	"io"
	"log"

	"github.com/go-logr/logr"
	ilog "github.com/shengyanli1982/orbit/internal/log"
	"k8s.io/klog/v2"
)

type LogrLogger struct {
	// l 是内部 zap.Logger 的引用
	// l is a reference to the internal zap.Logger
	l logr.Logger
}

func NewLogrLogger(w io.Writer) *LogrLogger {
	if w != nil {
		klog.LogToStderr(false) // 禁止输出到 stderr
		klog.SetOutput(w)       // 再次确保输出到我们的 writer
	}

	return &LogrLogger{
		l: klog.NewKlogr(),
	}
}

func (k *LogrLogger) GetLogrLogger() logr.Logger {
	return k.l
}

func (k *LogrLogger) GetStdLogger() *log.Logger {
	return ilog.NewStdLoggerFromLogr(&k.l)
}
