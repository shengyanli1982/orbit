package log

import (
	"log"

	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit/internal/conver"
)

// LogrAdapter 结构体实现了 io.Writer 接口，用于将标准日志适配到 logr.Logger。
// The LogrAdapter struct implements the io.Writer interface to adapt standard logs to logr.Logger.
type LogrAdapter struct {
	logger *logr.Logger // logr 日志记录器实例 (logr logger instance)
}

// Write 方法实现了 io.Writer 接口，将字节切片转换为字符串并写入日志。
// The Write method implements the io.Writer interface, converting byte slices to strings and writing them to the log.
func (l *LogrAdapter) Write(p []byte) (n int, err error) {
	l.logger.Info(conver.BytesToString(p))
	return len(p), nil
}

// NewStandardLoggerFromLogr 函数创建一个新的标准日志记录器，该记录器使用 logr.Logger 作为底层写入器。
// The NewStandardLoggerFromLogr function creates a new standard logger that uses logr.Logger as the underlying writer.
func NewStandardLoggerFromLogr(logger *logr.Logger) *log.Logger {
	return log.New(&LogrAdapter{logger: logger}, "", 0)
}
