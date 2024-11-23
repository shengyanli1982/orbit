package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Deprecated: Use ZapLogger instead (since v0.2.9). This method will be removed in the next release.
type Logger = ZapLogger

// Deprecated: Use NewLogger instead (since v0.2.9). This method will be removed in the next release.
func NewLogger(ws zapcore.WriteSyncer, opts ...zap.Option) *Logger {
	return NewZapLogger(ws, opts...)
}
