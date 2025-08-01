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

// 默认日志记录器的名称
const DefaultLoggerName = "default"

// 返回一个自定义的 JSON 编码器
func UseJSONReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	// 禁用 HTML 转义，以提高日志可读性
	enc.SetEscapeHTML(false)
	return enc
}

// 日志编码配置
var LogEncodingConfig = zapcore.EncoderConfig{
	TimeKey:             "time",                        // 时间字段的键名
	LevelKey:            "level",                       // 日志级别的键名
	NameKey:             "logger",                      // 日志记录器名称的键名
	CallerKey:           "caller",                      // 调用者信息的键名
	FunctionKey:         zapcore.OmitKey,               // 忽略函数名称
	MessageKey:          "message",                     // 消息内容的键名
	StacktraceKey:       "stacktrace",                  // 堆栈跟踪的键名
	LineEnding:          zapcore.DefaultLineEnding,     // 使用默认的行结束符
	EncodeLevel:         zapcore.CapitalLevelEncoder,   // 使用大写字母编码日志级别
	EncodeTime:          zapcore.ISO8601TimeEncoder,    // 使用 ISO8601 格式编码时间
	EncodeDuration:      zapcore.StringDurationEncoder, // 将持续时间编码为字符串
	EncodeCaller:        zapcore.ShortCallerEncoder,    // 使用短格式编码调用者信息
	NewReflectedEncoder: UseJSONReflectedEncoder,       // 使用自定义的 JSON 编码器
}

// 封装了多种类型的日志记录器
type ZapLogger struct {
	l   *zap.Logger        // Zap 日志记录器
	rl  *logr.Logger       // Logr 接口日志记录器
	sl  *log.Logger        // 标准库日志记录器
	sul *zap.SugaredLogger // Zap 语法糖日志记录器
}

// 创建并返回一个新的 ZapLogger 实例
func NewZapLogger(ws zapcore.WriteSyncer, isRelease bool, opts ...zap.Option) *ZapLogger {
	// 如果没有提供 WriteSyncer，则使用标准输出
	if ws == nil {
		ws = zapcore.AddSync(os.Stdout)
	}

	// 根据运行模式设置日志级别
	logLevel := zap.DebugLevel
	if isRelease {
		logLevel = zap.InfoLevel
	}

	// 创建核心日志组件
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(LogEncodingConfig),
		ws,
		zap.NewAtomicLevelAt(logLevel),
	)

	// 初始化各种日志记录器
	l := zap.New(core, zap.AddCaller()).WithOptions(opts...)
	rl := zapr.NewLogger(l)
	return &ZapLogger{l: l, rl: &rl, sul: l.Sugar(), sl: zap.NewStdLog(l)}
}

// 根据运行模式创建合适的 ZapLogger 实例的便利函数
// release 模式：输出 Info 级别以上的日志
// debug 模式：输出 Debug 级别以上的日志
func NewZapLoggerWithMode(ws zapcore.WriteSyncer, isReleaseMode bool, opts ...zap.Option) *ZapLogger {
	return NewZapLogger(ws, isReleaseMode, opts...)
}

// 返回原始的 Zap 日志记录器
func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.l
}

// 返回 Zap 语法糖日志记录器
func (l *ZapLogger) GetZapSugaredLogger() *zap.SugaredLogger {
	return l.sul
}

// 返回标准库日志记录器
func (l *ZapLogger) GetStandardLogger() *log.Logger {
	return l.sl
}

// 返回 Logr 接口日志记录器
func (l *ZapLogger) GetLogrLogger() *logr.Logger {
	return l.rl
}
