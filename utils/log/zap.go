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

// UseJSONReflectedEncoder 返回一个使用 json 解析器的 zapcore.ReflectedEncoder
// UseJSONReflectedEncoder returns a zapcore.ReflectedEncoder using json parser
func UseJSONReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	// 创建一个新的 json 编码器
	// Create a new json encoder
	enc := json.NewEncoder(w)

	// 设置编码器不转义 HTML 字符
	// Set the encoder not to escape HTML characters
	enc.SetEscapeHTML(false)

	// 返回编码器
	// Return the encoder
	return enc
}

// LogEncodingConfig 表示日志编码配置
// LogEncodingConfig represents the log encoding configuration
var LogEncodingConfig = zapcore.EncoderConfig{
	// TimeKey 是时间字段的键名
	// TimeKey is the key name of the time field
	TimeKey: "time",

	// LevelKey 是级别字段的键名
	// LevelKey is the key name of the level field
	LevelKey: "level",

	// NameKey 是记录器名称字段的键名
	// NameKey is the key name of the logger name field
	NameKey: "logger",

	// CallerKey 是调用者信息字段的键名
	// CallerKey is the key name of the caller information field
	CallerKey: "caller",

	// FunctionKey 是函数信息字段的键名，这里设置为 OmitKey，表示忽略该字段
	// FunctionKey is the key name of the function information field, here set to OmitKey, which means to ignore this field
	FunctionKey: zapcore.OmitKey,

	// MessageKey 是消息字段的键名
	// MessageKey is the key name of the message field
	MessageKey: "message",

	// StacktraceKey 是堆栈跟踪字段的键名
	// StacktraceKey is the key name of the stacktrace field
	StacktraceKey: "stacktrace",

	// LineEnding 是行结束符，这里使用默认的行结束符
	// LineEnding is the line ending character, here using the default line ending character
	LineEnding: zapcore.DefaultLineEnding,

	// EncodeLevel 是级别字段的编码器，这里使用 CapitalLevelEncoder，表示级别字段的值将被转换为大写
	// EncodeLevel is the encoder for the level field, here using CapitalLevelEncoder, which means the value of the level field will be converted to uppercase
	EncodeLevel: zapcore.CapitalLevelEncoder,

	// EncodeTime 是时间字段的编码器，这里使用 ISO8601TimeEncoder，表示时间字段的值将被转换为 ISO 8601 格式
	// EncodeTime is the encoder for the time field, here using ISO8601TimeEncoder, which means the value of the time field will be converted to ISO 8601 format
	EncodeTime: zapcore.ISO8601TimeEncoder,

	// EncodeDuration 是持续时间字段的编码器，这里使用 StringDurationEncoder，表示持续时间字段的值将被转换为字符串格式
	// EncodeDuration is the encoder for the duration field, here using StringDurationEncoder, which means the value of the duration field will be converted to string format
	EncodeDuration: zapcore.StringDurationEncoder,

	// EncodeCaller 是调用者信息字段的编码器，这里使用 ShortCallerEncoder，表示调用者信息字段的值将被转换为短格式
	// EncodeCaller is the encoder for the caller information field, here using ShortCallerEncoder, which means the value of the caller information field will be converted to short format
	EncodeCaller: zapcore.ShortCallerEncoder,

	// NewReflectedEncoder 是反射编码器的创建函数，这里使用 UseJSONReflectedEncoder，表示创建一个使用 json 解析器的反射编码器
	// NewReflectedEncoder is the creation function for the reflected encoder, here using UseJSONReflectedEncoder, which means to create a reflected encoder using a json parser
	NewReflectedEncoder: UseJSONReflectedEncoder,
}

// ZapLogger 结构体包装了 zap.ZapLogger
// The ZapLogger struct wraps zap.ZapLogger
type ZapLogger struct {
	// l 是内部 zap.Logger 的引用
	// l is a reference to the internal zap.Logger
	l *zap.Logger
}

// NewZapLogger 创建一个新的 Logger
// NewZapLogger creates a new Logger
func NewZapLogger(ws zapcore.WriteSyncer, opts ...zap.Option) *ZapLogger {
	// 如果 ws 为空，则默认使用 os.Stdout
	// If ws is nil, use os.Stdout by default
	if ws == nil {
		ws = zapcore.AddSync(os.Stdout)
	}

	// 创建一个新的 zapcore.Core
	// Create a new zapcore.Core
	core := zapcore.NewCore(
		// 使用 LogEncodingConfig 创建一个新的 JSON 编码器
		// Create a new JSON encoder using LogEncodingConfig
		zapcore.NewJSONEncoder(LogEncodingConfig),
		// 使用 ws 作为输出
		// Use ws as the output
		ws,
		// 设置日志级别为 Debug
		// Set the log level to Debug
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)

	// 返回一个新的 Logger，其中包含了一个 zap.Logger
	// Return a new Logger that contains a zap.Logger
	return &ZapLogger{l: zap.New(core, zap.AddCaller()).WithOptions(opts...)}
}

// GetZapLogger 返回 zap.Logger
// GetZapLogger returns the zap.Logger
func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.l
}

// GetZapSugaredLogger 返回 zap.SugaredLogger
// GetZapSugaredLogger returns the zap.SugaredLogger
func (l *ZapLogger) GetZapSugaredLogger() *zap.SugaredLogger {
	return l.l.Sugar()
}

// GetStandardLogger 返回标准库的 logger
// GetStandardLogger returns the standard library logger
func (l *ZapLogger) GetStandardLogger() *log.Logger {
	return zap.NewStdLog(l.l)
}

// GetLogrLogger 返回 logr.Logger
// GetLogrLogger returns the logr.Logger
func (l *ZapLogger) GetLogrLogger() logr.Logger {
	return zapr.NewLogger(l.l)
}
