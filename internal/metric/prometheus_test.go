package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServerMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.requestCount)
	assert.NotNil(t, metrics.requestLatencies)
	assert.NotNil(t, metrics.requestLatency)
	assert.Equal(t, registry, metrics.registry)
}

func TestServerMetricsHandlerFunc(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	// Create a test router
	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	// Define a test route
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response")
	})

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	m := &dto.Metric{}

	// Assert the metrics
	_ = metrics.requestCount.WithLabelValues("GET", "/test", "200").Write(m)
	assert.Equal(t, 1, int(m.Counter.GetValue()))
	_ = metrics.requestLatency.WithLabelValues("GET", "/test", "200").Write(m)
	assert.Equal(t, 1, int(m.Counter.GetValue()))
}
