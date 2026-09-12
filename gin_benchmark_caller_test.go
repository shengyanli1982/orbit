package orbit

import (
	"io"
	"testing"

	"github.com/go-logr/logr"
	ulog "github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap/zapcore"
)

// newBenchmarkDiscardLoggerNoCaller 返回写入 io.Discard 且关闭 caller 的 logr 日志记录器：
// 与 newBenchmarkDiscardLogger 唯一差异是 WithLogCaller(false)（跳过每条日志的调用栈捕获），
// 用于量化 AddCaller 在真实日志路径上的开销
func newBenchmarkDiscardLoggerNoCaller() *logr.Logger {
	return ulog.NewZapLogger(zapcore.AddSync(io.Discard), true, ulog.WithLogCaller(false)).GetLogrLogger()
}

// BenchmarkEngineMainPathRealLogNoCaller 与 BenchmarkEngineMainPathRealLog 同构，
// 仅日志器关闭 caller（WithLogCaller(false)），用于实证 AddCaller 的
// 栈捕获 + caller 编码开销（预期约 -100ns/op、-1 alloc/op）
func BenchmarkEngineMainPathRealLogNoCaller(b *testing.B) {
	benchmarkEngineMainPathRealLog(b, newBenchmarkDiscardLoggerNoCaller())
}
