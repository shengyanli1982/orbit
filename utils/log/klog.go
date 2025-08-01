package log

import (
	"flag"
	"io"
	"log"
	"os"
	"sync"

	"github.com/go-logr/logr"
	ilog "github.com/shengyanli1982/orbit/internal/log"
	"k8s.io/klog/v2"
)

// 用于确保 klog 标志只被初始化一次
var once sync.Once

// 初始化 klog 的标志配置
func initKlogFlags(isRelease bool) {
	once.Do(func() {
		// 创建一个新的标志集，用于 klog 配置
		fs := flag.NewFlagSet("klog", flag.PanicOnError)

		// 初始化 klog 标志
		klog.InitFlags(fs)

		// 设置 klog 的输出选项
		_ = fs.Set("one_output", "true")       // 启用单一输出
		_ = fs.Set("logtostderr", "false")     // 禁用标准错误输出
		_ = fs.Set("alsologtostderr", "false") // 禁用同时输出到标准错误
		_ = fs.Set("stderrthreshold", "FATAL") // 设置标准错误阈值为 FATAL

		// 根据运行模式设置日志详细程度
		if isRelease {
			// Release 模式: 降低详细程度，只输出 Info 级别以上的日志
			_ = fs.Set("v", "0")
		} else {
			// Debug 模式: 提高详细程度，输出更多调试信息
			_ = fs.Set("v", "2")
		}

		// 解析标志，不传入任何参数
		_ = fs.Parse(nil)
	})
}

// 包装了 logr.Logger 和标准日志记录器
type LogrLogger struct {
	l  logr.Logger // logr 日志记录器
	sl *log.Logger // 标准日志记录器
}

// 创建并返回一个新的 LogrLogger 实例
func NewLogrLogger(w io.Writer, isRelease bool) *LogrLogger {
	// 初始化 klog 标志
	initKlogFlags(isRelease)

	// 如果没有提供写入器，则使用标准输出
	if w == nil {
		w = os.Stdout
	}

	klog.SetOutput(w)  // 设置输出写入器
	klog.ClearLogger() // 清除现有的日志记录器

	// 创建新的 klog 记录器
	l := klog.NewKlogr()

	// 返回包装了 logr 和标准日志记录器的实例
	return &LogrLogger{
		l:  l,
		sl: ilog.NewStandardLoggerFromLogr(&l),
	}
}

// 根据运行模式创建合适的 LogrLogger 实例的便利函数
// release 模式：降低详细程度，只输出 Info 级别以上的日志 (verbosity=0)
// debug 模式：提高详细程度，输出更多调试信息 (verbosity=2)
func NewLogrLoggerWithMode(w io.Writer, isReleaseMode bool) *LogrLogger {
	return NewLogrLogger(w, isReleaseMode)
}

// 返回 logr.Logger 的指针
func (k *LogrLogger) GetLogrLogger() *logr.Logger {
	return &k.l
}

// 返回标准日志记录器的指针
func (k *LogrLogger) GetStandardLogger() *log.Logger {
	return k.sl
}
