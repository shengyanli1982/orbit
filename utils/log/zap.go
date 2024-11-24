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

// DefaultLoggerName 是默认日志记录器的名称。
// DefaultLoggerName is the name of the default logger.
const DefaultLoggerName = "default"

// UseJSONReflectedEncoder 函数返回一个自定义的 JSON 编码器。
// The UseJSONReflectedEncoder function returns a custom JSON encoder.
func UseJSONReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	// 禁用 HTML 转义，以提高日志可读性
	// Disable HTML escaping to improve log readability
	enc.SetEscapeHTML(false)
	return enc
}

// LogEncodingConfig 定义了日志编码的配置。
// LogEncodingConfig defines the configuration for log encoding.
var LogEncodingConfig = zapcore.EncoderConfig{
	TimeKey:             "time",                        // 时间字段的键名 (Key name for the time field)
	LevelKey:            "level",                       // 日志级别的键名 (Key name for the level field)
	NameKey:             "logger",                      // 日志记录器名称的键名 (Key name for the logger name)
	CallerKey:           "caller",                      // 调用者信息的键名 (Key name for the caller information)
	FunctionKey:         zapcore.OmitKey,               // 忽略函数名称 (Omit function name)
	MessageKey:          "message",                     // 消息内容的键名 (Key name for the message content)
	StacktraceKey:       "stacktrace",                  // 堆栈跟踪的键名 (Key name for the stack trace)
	LineEnding:          zapcore.DefaultLineEnding,     // 使用默认的行结束符 (Use default line ending)
	EncodeLevel:         zapcore.CapitalLevelEncoder,   // 使用大写字母编码日志级别 (Encode log level in capital letters)
	EncodeTime:          zapcore.ISO8601TimeEncoder,    // 使用 ISO8601 格式编码时间 (Encode time in ISO8601 format)
	EncodeDuration:      zapcore.StringDurationEncoder, // 将持续时间编码为字符串 (Encode duration as string)
	EncodeCaller:        zapcore.ShortCallerEncoder,    // 使用短格式编码调用者信息 (Encode caller information in short format)
	NewReflectedEncoder: UseJSONReflectedEncoder,       // 使用自定义的 JSON 编码器 (Use custom JSON encoder)
}

// ZapLogger 结构体封装了多种类型的日志记录器。
// The ZapLogger struct encapsulates multiple types of loggers.
type ZapLogger struct {
	l   *zap.Logger        // Zap 日志记录器 (Zap logger)
	rl  *logr.Logger       // Logr 接口日志记录器 (Logr interface logger)
	sl  *log.Logger        // 标准库日志记录器 (Standard library logger)
	sul *zap.SugaredLogger // Zap 语法糖日志记录器 (Zap sugared logger)
}

// NewZapLogger 函数创建并返回一个新的 ZapLogger 实例。
// The NewZapLogger function creates and returns a new ZapLogger instance.
func NewZapLogger(ws zapcore.WriteSyncer, opts ...zap.Option) *ZapLogger {
	// 如果没有提供 WriteSyncer，则使用标准输出
	// If no WriteSyncer is provided, use standard output
	if ws == nil {
		ws = zapcore.AddSync(os.Stdout)
	}

	// 创建核心日志组件
	// Create core logging component
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(LogEncodingConfig),
		ws,
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)

	// 初始化各种日志记录器
	// Initialize various loggers
	l := zap.New(core, zap.AddCaller()).WithOptions(opts...)
	rl := zapr.NewLogger(l)
	return &ZapLogger{l: l, rl: &rl, sul: l.Sugar(), sl: zap.NewStdLog(l)}
}

// GetZapLogger 方法返回原始的 Zap 日志记录器。
// The GetZapLogger method returns the original Zap logger.
func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.l
}

// GetZapSugaredLogger 方法返回 Zap 语法糖日志记录器。
// The GetZapSugaredLogger method returns the Zap sugared logger.
func (l *ZapLogger) GetZapSugaredLogger() *zap.SugaredLogger {
	return l.sul
}

// GetStandardLogger 方法返回标准库日志记录器。
// The GetStandardLogger method returns the standard library logger.
func (l *ZapLogger) GetStandardLogger() *log.Logger {
	return l.sl
}

// GetLogrLogger 方法返回 Logr 接口日志记录器。
// The GetLogrLogger method returns the Logr interface logger.
func (l *ZapLogger) GetLogrLogger() *logr.Logger {
	return l.rl
}
