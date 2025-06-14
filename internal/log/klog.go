package log

import (
	"log"

	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit/internal/conver"
)

// 实现了 io.Writer 接口，用于将标准日志适配到 logr.Logger
type LogrAdapter struct {
	logger *logr.Logger // logr 日志记录器实例
}

// 实现了 io.Writer 接口，将字节切片转换为字符串并写入日志
func (l *LogrAdapter) Write(p []byte) (n int, err error) {
	l.logger.Info(conver.BytesToString(p))
	return len(p), nil
}

// 创建一个新的标准日志记录器，该记录器使用 logr.Logger 作为底层写入器
func NewStandardLoggerFromLogr(logger *logr.Logger) *log.Logger {
	return log.New(&LogrAdapter{logger: logger}, "", 0)
}
