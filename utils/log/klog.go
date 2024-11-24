package log

import (
	"flag"
	"io"
	"log"
	"sync"

	"github.com/go-logr/logr"
	ilog "github.com/shengyanli1982/orbit/internal/log"
	"k8s.io/klog/v2"
)

// once 是一个同步器，用于确保 klog 标志只被初始化一次。
// once is a synchronizer to ensure klog flags are initialized only once.
var once sync.Once

// initKlogFlags 函数初始化 klog 的标志配置。
// The initKlogFlags function initializes the klog flags configuration.
func initKlogFlags() {
	once.Do(func() {
		// 创建一个新的标志集，用于 klog 配置
		// Create a new flag set for klog configuration
		fs := flag.NewFlagSet("klog", flag.PanicOnError)
		
		// 初始化 klog 标志
		// Initialize klog flags
		klog.InitFlags(fs)
		
		// 设置 klog 的输出选项
		// Set klog output options
		_ = fs.Set("one_output", "true")      // 启用单一输出 (Enable single output)
		_ = fs.Set("logtostderr", "false")    // 禁用标准错误输出 (Disable logging to stderr)
		_ = fs.Set("alsologtostderr", "false") // 禁用同时输出到标准错误 (Disable also logging to stderr)
		_ = fs.Set("stderrthreshold", "FATAL") // 设置标准错误阈值为 FATAL (Set stderr threshold to FATAL)
		
		// 解析标志，不传入任何参数
		// Parse flags without any arguments
		_ = fs.Parse(nil)
	})
}

// LogrLogger 结构体包装了 logr.Logger 和标准日志记录器。
// The LogrLogger struct wraps logr.Logger and standard logger.
type LogrLogger struct {
	l  logr.Logger  // logr 日志记录器 (logr logger)
	sl *log.Logger  // 标准日志记录器 (standard logger)
}

// NewLogrLogger 函数创建并返回一个新的 LogrLogger 实例。
// The NewLogrLogger function creates and returns a new LogrLogger instance.
func NewLogrLogger(w io.Writer) *LogrLogger {
	// 初始化 klog 标志
	// Initialize klog flags
	initKlogFlags()

	// 如果提供了写入器，则配置 klog 输出
	// Configure klog output if writer is provided
	if w != nil {
		klog.SetOutput(w)    // 设置输出写入器 (Set output writer)
		klog.ClearLogger()   // 清除现有的日志记录器 (Clear existing logger)
	}

	// 创建新的 klog 记录器
	// Create new klog logger
	l := klog.NewKlogr()
	
	// 返回包装了 logr 和标准日志记录器的实例
	// Return instance wrapping both logr and standard logger
	return &LogrLogger{
		l:  l,
		sl: ilog.NewStandardLoggerFromLogr(&l),
	}
}

// GetLogrLogger 方法返回 logr.Logger 的指针。
// The GetLogrLogger method returns a pointer to the logr.Logger.
func (k *LogrLogger) GetLogrLogger() *logr.Logger {
	return &k.l
}

// GetStandardLogger 方法返回标准日志记录器的指针。
// The GetStandardLogger method returns a pointer to the standard logger.
func (k *LogrLogger) GetStandardLogger() *log.Logger {
	return k.sl
}
