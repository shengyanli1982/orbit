package metric

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestServerMetricsHandlerFuncDoesNotLogContextErrors 验证 #8：metric 中间件不承担
// 错误日志职责——context.Errors 由 AccessLogger 统一记录一次，避免 ReleaseOptions
// （默认启用 metric）下每条请求错误被双重打日志；指标采集职责（计数/时延）不受影响。
func TestServerMetricsHandlerFuncDoesNotLogContextErrors(t *testing.T) {
	core, recorded := observer.New(zapcore.ErrorLevel)
	logger := zapr.NewLogger(zap.New(core))

	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))
	router.GET("/test", func(c *gin.Context) {
		_ = c.Error(errors.New("boom"))
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 0, recorded.FilterMessage("Error occurred").Len(),
		"metric middleware must not log context.Errors (AccessLogger owns error logging)")

	// 指标采集职责不受影响：请求计数仍正常记录
	m := &dto.Metric{}
	require.NoError(t, metrics.requestCount.WithLabelValues("GET", "/test", "200").Write(m))
	assert.Equal(t, 1, int(m.Counter.GetValue()))
}
