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

// DefaultLoggerName 是默认的日志记录器名称
// DefaultLoggerName is the default logger name
const DefaultLoggerName = "default"

// UseJSONReflectedEncoder 返回使用 json 解析器的 zapcore.ReflectedEncoder
// UseJSONReflectedEncoder returns a zapcore.ReflectedEncoder using json parser
func UseJSONReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

// LogEncodingConfig 表示日志编码配置
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

// NewLogger 创建一个新的日志记录器
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

// GetZapLogger 返回 zap 日志记录器
// GetZapLogger returns the zap logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.l
}

// GetZapSugaredLogger 返回 zap 日志记录器 sugar
// GetZapSugaredLogger returns the zap logger sugar
func (l *Logger) GetZapSugaredLogger() *zap.SugaredLogger {
	return l.l.Sugar()
}

// GetStdLogger 返回标准库日志记录器
// GetStdLogger returns the standard library logger
func (l *Logger) GetStdLogger() *log.Logger {
	return zap.NewStdLog(l.l)
}

// GetLogrLogger 返回 logr 日志记录器
// GetLogrLogger returns the logr logger
func (l *Logger) GetLogrLogger() logr.Logger {
	return zapr.NewLogger(l.l)
}
