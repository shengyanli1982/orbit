package log

import (
	"log"

	ilog "github.com/shengyanli1982/orbit/internal/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Deprecated: Use ZapLogger instead (since v0.2.9). This method will be removed in the next release.
type Logger = ZapLogger

// Deprecated: Use NewLogger instead (since v0.2.9). This method will be removed in the next release.
func NewLogger(ws zapcore.WriteSyncer, opts ...zap.Option) *Logger {
	return NewZapLogger(ws, false, opts...)
}

// Deprecated: Use GetStandardLogger instead (since v0.2.9). This method will be removed in the next release.
func (l *ZapLogger) GetStdLogger() *log.Logger {
	return zap.NewStdLog(l.l)
}

// Deprecated: Use GetStandardLogger instead (since v0.2.9). This method will be removed in the next release.
func (k *LogrLogger) GetStdLogger() *log.Logger {
	return ilog.NewStandardLoggerFromLogr(&k.l)
}
