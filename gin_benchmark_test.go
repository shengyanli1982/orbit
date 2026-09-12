// 基准测量口径说明：httptest.ResponseRecorder.WriteHeader 每次都会克隆响应头
// （约 3 allocs / 488B per op，头数量越多成本越高），属测试夹具伪影——生产环境
// net/http 服务端不做该克隆。因此本文件各基准的绝对值系统性高于生产等效值，
// 不可直接当作生产数字；但夹具成本恒定，基准之间的差分对比仍然有效。
// 详见 .agent-work-b652f803/reports/pprof-analysis-round1.md §3.1。

package orbit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	ulog "github.com/shengyanli1982/orbit/utils/log"
)

type benchmarkService struct{}

func (s *benchmarkService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/bench/:id", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})
}

func benchmarkNoopLogEvent(_ *logr.Logger, _ *ulog.LogEvent) {}

func newBenchmarkEngine(tb testing.TB, enableMetric bool) *Engine {
	tb.Helper()

	options := NewOptions()
	if enableMetric {
		options = options.EnableMetric()
	}

	// 复用静音配置（日志导向 io.Discard）：消除 engine.Run 生命周期 INFO 日志
	// （http server is ready/shutdown）对 bench 输出行的撕裂与噪声抬升；
	// 测量语义不变——仍走真实 Run() 装配路径 + httptest 进程内请求
	return newBenchmarkEngineWith(tb, newQuietBenchmarkConfig(), options, &benchmarkService{})
}

func TestBenchmarkEngineMainPathSetup(t *testing.T) {
	engine := newBenchmarkEngine(t, false)
	req := httptest.NewRequest(http.MethodGet, "/bench/123?foo=bar", nil)
	resp := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
	}
}

func BenchmarkEngineMainPath(b *testing.B) {
	cases := []struct {
		name         string
		enableMetric bool
		withOrigin   bool
	}{
		{name: "NoMetric_NoOrigin", enableMetric: false, withOrigin: false},
		{name: "NoMetric_WithOrigin", enableMetric: false, withOrigin: true},
		{name: "Metric_NoOrigin", enableMetric: true, withOrigin: false},
		{name: "Metric_WithOrigin", enableMetric: true, withOrigin: true},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			engine := newBenchmarkEngine(b, tc.enableMetric)

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				req := httptest.NewRequest(http.MethodGet, "/bench/123?foo=bar", nil)
				if tc.withOrigin {
					req.Header.Set("Origin", "https://app.example.com")
				}

				for pb.Next() {
					resp := httptest.NewRecorder()
					engine.ginSvr.ServeHTTP(resp, req)
					if resp.Code != http.StatusOK {
						b.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
					}
				}
			})
		})
	}
}

func BenchmarkEngineMainPathMiss(b *testing.B) {
	cases := []struct {
		name       string
		method     string
		path       string
		expectCode int
	}{
		{name: "NotFound_Metric", method: http.MethodGet, path: "/not-found?foo=bar", expectCode: http.StatusNotFound},
		{name: "MethodNotAllowed_Metric", method: http.MethodPost, path: "/bench/123", expectCode: http.StatusMethodNotAllowed},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			engine := newBenchmarkEngine(b, true)

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				req := httptest.NewRequest(tc.method, tc.path, nil)

				for pb.Next() {
					resp := httptest.NewRecorder()
					engine.ginSvr.ServeHTTP(resp, req)
					if resp.Code != tc.expectCode {
						b.Fatalf("unexpected status code: %d, body: %s", resp.Code, resp.Body.String())
					}
				}
			})
		})
	}
}
