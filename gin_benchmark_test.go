package orbit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	mid "github.com/shengyanli1982/orbit/internal/middleware"
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

	config := NewConfig().
		WithRelease().
		WithAccessLogEventFunc(benchmarkNoopLogEvent).
		WithRecoveryLogEventFunc(benchmarkNoopLogEvent).
		WithPrometheusRegistry(prometheus.NewRegistry())

	options := NewOptions()
	if enableMetric {
		options = options.EnableMetric()
	}

	engine := NewEngine(config, options)
	if engine.initErr != nil {
		tb.Fatalf("engine init failed: %v", engine.initErr)
	}

	engine.RegisterService(&benchmarkService{})
	engine.registerUserMiddlewares()
	engine.ginSvr.Use(mid.AccessLogger(engine.config.logger, engine.config.accessLogEventFunc, engine.opts.recReqBody))
	engine.registerUserServices()
	return engine
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
		tc := tc
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
