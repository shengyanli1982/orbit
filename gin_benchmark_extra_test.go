package orbit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	ulog "github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap/zapcore"
)

// newBenchmarkDiscardLogger 返回写入 io.Discard 的 logr 日志记录器。
// 使用与生产默认日志器同构的 NewZapLogger 设施（JSON 编码器、Info 级别、带 caller），
// 仅写出目标为 io.Discard——用于测量日志序列化/编码开销而非磁盘 I/O
func newBenchmarkDiscardLogger() *logr.Logger {
	return ulog.NewZapLogger(zapcore.AddSync(io.Discard), true).GetLogrLogger()
}

// newQuietBenchmarkConfig 返回基准测试用配置：Release 模式、日志全部导向
// io.Discard、访问/恢复事件函数为 no-op，消除生命周期与请求日志的输出噪音
func newQuietBenchmarkConfig() *Config {
	return NewConfig().
		WithRelease().
		WithLogger(newBenchmarkDiscardLogger()).
		WithAccessLogEventFunc(benchmarkNoopLogEvent).
		WithRecoveryLogEventFunc(benchmarkNoopLogEvent)
}

// newBenchmarkEngineWith 按给定配置与选项装配基准引擎。
// 中间件链一律经 Engine/Run 装配（不直接调用 middleware 构造函数），
// 对中间件构造签名的变化免疫
func newBenchmarkEngineWith(tb testing.TB, config *Config, options *Options, services ...Service) *Engine {
	tb.Helper()

	config = config.
		WithPort(getFreePort(tb)).
		WithPrometheusRegistry(prometheus.NewRegistry())

	engine := NewEngine(config, options)
	if engine.initErr != nil {
		tb.Fatalf("engine init failed: %v", engine.initErr)
	}

	for _, service := range services {
		engine.RegisterService(service)
	}
	engine.Run()
	tb.Cleanup(engine.Stop)
	return engine
}

// benchmarkEngineMainPathRealLog 是 RealLog 系基准的共享测量体：
// accessLogEventFunc 使用 DefaultAccessEventFunc（同步消费），日志器由参数指定，
// 其余请求与中间件链同 BenchmarkEngineMainPath/Metric_WithOrigin（仅日志消费方式不同），
// 保证变体之间除日志器外完全同构、可直接差分对比
func benchmarkEngineMainPathRealLog(b *testing.B, logger *logr.Logger) {
	b.Helper()

	config := newQuietBenchmarkConfig().
		WithLogger(logger).
		WithAccessLogEventFunc(ulog.DefaultAccessEventFunc)
	engine := newBenchmarkEngineWith(b, config, NewOptions().EnableMetric(), &benchmarkService{})

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/bench/123?foo=bar", nil)
		req.Header.Set("Origin", "https://app.example.com")

		for pb.Next() {
			resp := httptest.NewRecorder()
			engine.ginSvr.ServeHTTP(resp, req)
			if resp.Code != http.StatusOK {
				b.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
			}
		}
	})
}

// BenchmarkEngineMainPathRealLog 测量真实访问日志路径的开销：
// 日志器为与生产同构的 zap 设施（带 caller）、写出到 io.Discard——
// JSON 序列化与编码开销计入测量，磁盘 I/O 不计
func BenchmarkEngineMainPathRealLog(b *testing.B) {
	benchmarkEngineMainPathRealLog(b, newBenchmarkDiscardLogger())
}

// BenchmarkEngineCORSOptions 测量 CORS 预检路径：OPTIONS + Origin +
// Access-Control-Request-Method，由 CORS 中间件短路中止（不进入业务 handler）。
// 预期状态码以计时前的一次探测为准（当前行为为 204），测量循环内断言一致性，
// 避免与 CORS 预检行为的调整强耦合
func BenchmarkEngineCORSOptions(b *testing.B) {
	engine := newBenchmarkEngineWith(b, newQuietBenchmarkConfig(), NewOptions().EnableMetric(), &benchmarkService{})

	newPreflightRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodOptions, "/bench/123", nil)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		return req
	}

	probe := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(probe, newPreflightRequest())
	expectCode := probe.Code
	if expectCode < 200 || expectCode >= 300 {
		b.Fatalf("unexpected preflight status code: %d, body: %s", expectCode, probe.Body.String())
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := newPreflightRequest()

		for pb.Next() {
			resp := httptest.NewRecorder()
			engine.ginSvr.ServeHTTP(resp, req)
			if resp.Code != expectCode {
				b.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
			}
		}
	})
}

// benchmarkPanicService 注册必然 panic 的路由，用于测量 Recovery 恢复路径
type benchmarkPanicService struct{}

func (s *benchmarkPanicService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/panic", func(ctx *gin.Context) {
		panic("benchmark panic")
	})
}

// BenchmarkEngineRecoveryPanic 测量 panic 恢复路径：handler panic →
// Recovery 捕获（含 debug.Stack 栈采集）→ 500。
// 恢复事件函数为 no-op、日志导向 io.Discard，无输出噪音
func BenchmarkEngineRecoveryPanic(b *testing.B) {
	engine := newBenchmarkEngineWith(b, newQuietBenchmarkConfig(), NewOptions().EnableMetric(), &benchmarkPanicService{})

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)

		for pb.Next() {
			resp := httptest.NewRecorder()
			engine.ginSvr.ServeHTTP(resp, req)
			if resp.Code != http.StatusInternalServerError {
				b.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
			}
		}
	})
}

// BenchmarkEngineBodyBuffer 测量响应体缓冲路径：EnableRecordResponseBody 开启
// BodyBuffer 中间件，handler 返回普通大小响应，缓冲写入器的包装/双写/回收开销计入测量
func BenchmarkEngineBodyBuffer(b *testing.B) {
	options := NewOptions().EnableMetric().EnableRecordResponseBody()
	engine := newBenchmarkEngineWith(b, newQuietBenchmarkConfig(), options, &benchmarkService{})

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/bench/123?foo=bar", nil)

		for pb.Next() {
			resp := httptest.NewRecorder()
			engine.ginSvr.ServeHTTP(resp, req)
			if resp.Code != http.StatusOK {
				b.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
			}
			if body := resp.Body.String(); body != "OK" {
				b.Fatalf("unexpected body: %q", body)
			}
		}
	})
}
