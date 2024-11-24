package log

import (
	"log"

	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit/internal/conver"
)

type LogrAdapter struct {
	logger *logr.Logger
}

// 实现 io.Writer 接口，这样就可以用于创建 log.Logger
func (l *LogrAdapter) Write(p []byte) (n int, err error) {
	// // 去掉可能的换行符
	// msg := strings.TrimSuffix(string(p), "\n")
	l.logger.Info(conver.BytesToString(p))
	return len(p), nil
}

// 创建一个将 logr.Logger 转换为 log.Logger 的函数
func NewStandardLoggerFromLogr(logger *logr.Logger) *log.Logger {
	return log.New(&LogrAdapter{logger: logger}, "", 0) // 标志位设为0因为 logr 通常已经处理了时间戳等信息
}
