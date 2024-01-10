package log

import (
	"io"
	"log"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/shengyanli1982/orbit/internal/codec/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DefaultLoggerName = "default"

// UseJSONReflectedEncoder returns a zapcore.ReflectedEncoder using json parser
func UseJSONReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

// LogEncodingConfig represents the log encoding configuration
var LogEncodingConfig = zapcore.EncoderConfig{
	TimeKey:             "time",
	LevelKey:            "level",
	NameKey:             "logger",
	CallerKey:           "caller",
	FunctionKey:         zapcore.OmitKey,
	MessageKey:          "message",
	StacktraceKey:       "stacktrace",
	LineEnding:          zapcore.DefaultLineEnding,
	EncodeLevel:         zapcore.CapitalLevelEncoder,
	EncodeTime:          zapcore.ISO8601TimeEncoder,
	EncodeDuration:      zapcore.StringDurationEncoder,
	EncodeCaller:        zapcore.ShortCallerEncoder,
	NewReflectedEncoder: UseJSONReflectedEncoder,
}

type Logger struct {
	l *zap.Logger
}

// NewLogger creates a new logger
func NewLogger(ws zapcore.WriteSyncer, opts ...zap.Option) *Logger {
	if ws == nil {
		ws = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(LogEncodingConfig),
		ws,
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)

	return &Logger{l: zap.New(core, zap.AddCaller()).WithOptions(opts...)}
}

// GetZapLogger returns the zap logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.l
}

// GetZapSugaredLogger returns the zap logger sugar
func (l *Logger) GetZapSugaredLogger() *zap.SugaredLogger {
	return l.l.Sugar()
}

// GetStdLogger returns the standard library logger
func (l *Logger) GetStdLogger() *log.Logger {
	return zap.NewStdLog(l.l)
}

// GetLogrLogger returns the logr logger
func (l *Logger) GetLogrLogger() logr.Logger {
	return zapr.NewLogger(l.l)
}
