package orbit

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

var (
	defaultLoggerName = "default"
)

// 根据 build 参数定义使用那种 json 解析器
// Use json parser based on build parameters
func jsonReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

// 日志编码配置
// Log encoding configuration
var enccfg = zapcore.EncoderConfig{
	TimeKey:             "time",
	LevelKey:            "level",
	NameKey:             "logger",
	CallerKey:           "caller",
	FunctionKey:         zapcore.OmitKey,
	MessageKey:          "message",
	StacktraceKey:       "stacktrace",
	LineEnding:          zapcore.DefaultLineEnding,
	EncodeLevel:         zapcore.CapitalLevelEncoder, // 日志级别使用大写显示 (Log level is displayed in uppercase)
	EncodeTime:          zapcore.ISO8601TimeEncoder,  // 自定义输出时间格式 (Custom output time format)
	EncodeDuration:      zapcore.StringDurationEncoder,
	EncodeCaller:        zapcore.ShortCallerEncoder,
	NewReflectedEncoder: jsonReflectedEncoder, // 使用 json 解析器 (Use json parser)
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
		zapcore.NewJSONEncoder(enccfg), // 构建 json 解码器 (Build json decoder)
		ws,
		zap.NewAtomicLevelAt(zap.DebugLevel), // 将日志级别设置为 DEBUG (Set log level to DEBUG)
	)

	return &Logger{l: zap.New(core, zap.AddCaller()).WithOptions(opts...)}
}

// 返回 zap 日志记录器
// Return zap logger
func (l *Logger) L() *zap.Logger {
	return l.l
}

// 返回 zap 日志记录器的 sugar
// Return zap logger sugar
func (l *Logger) S() *zap.SugaredLogger {
	return l.l.Sugar()
}

// 返回标准库日志记录器
// Return standard library logger
func (l *Logger) Std() *log.Logger {
	return zap.NewStdLog(l.l)
}

// 返回 logr 日志记录器
// Return logr logger
func (l *Logger) Lr() logr.Logger {
	return zapr.NewLogger(l.l)
}
