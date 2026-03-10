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
	"github.com/stretchr/testify/require"
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
	g := &dto.Metric{}
	_ = metrics.requestLatency.WithLabelValues("GET", "/test", "200").Write(g)
	assert.NotNil(t, g.GetGauge())
	assert.GreaterOrEqual(t, g.GetGauge().GetValue(), 0.0)
}

func TestServerMetricsPrometheusNaming(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	metrics.IncRequestCount("GET", "/test", "200")
	metrics.ObserveRequestLatency("GET", "/test", "200", 0.2)
	metrics.SetRequestLatency("GET", "/test", "200", 0.2)
	metrics.Register()
	defer metrics.Unregister()

	families, err := registry.Gather()
	assert.NoError(t, err)

	names := make(map[string]struct{}, len(families))
	for _, family := range families {
		names[family.GetName()] = struct{}{}
	}

	assert.Contains(t, names, "orbit_http_requests_total")
	assert.Contains(t, names, "orbit_http_request_duration_seconds")
	assert.Contains(t, names, "orbit_http_request_duration_seconds_last")
}

func TestServerMetricsDurationBucketsCover120Seconds(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	metrics.ObserveRequestLatency("GET", "/test", "200", 95)
	metrics.Register()
	defer metrics.Unregister()

	families, err := registry.Gather()
	require.NoError(t, err)

	var durationFamily *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "orbit_http_request_duration_seconds" {
			durationFamily = family
			break
		}
	}
	require.NotNil(t, durationFamily)
	require.NotEmpty(t, durationFamily.Metric)

	histogram := durationFamily.Metric[0].GetHistogram()
	require.NotNil(t, histogram)

	has120sBucket := false
	maxUpperBound := 0.0
	for _, bucket := range histogram.Bucket {
		upper := bucket.GetUpperBound()
		if upper > maxUpperBound {
			maxUpperBound = upper
		}
		if upper == 120 {
			has120sBucket = true
		}
	}

	assert.True(t, has120sBucket, "duration histogram should include 120s bucket")
	assert.GreaterOrEqual(t, maxUpperBound, 120.0)
}

func TestServerMetricsPathNormalization(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	// Create a test router
	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	// Define test routes with parameters
	router.GET("/users/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "User response")
	})
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response")
	})

	t.Run("Registered route uses route template", func(t *testing.T) {
		// Make request to parametrized route
		req, _ := http.NewRequest("GET", "/users/123", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		m := &dto.Metric{}
		// Should use route template /users/:id, not /users/123
		_ = metrics.requestCount.WithLabelValues("GET", "/users/:id", "200").Write(m)
		assert.Equal(t, 1, int(m.Counter.GetValue()), "Should track route template")
	})

	t.Run("Unregistered route returns unmatched", func(t *testing.T) {
		// Make request to unregistered route
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		m := &dto.Metric{}
		// Should use "unmatched" for unregistered routes
		_ = metrics.requestCount.WithLabelValues("GET", "unmatched", "404").Write(m)
		assert.Equal(t, 1, int(m.Counter.GetValue()), "Should use unmatched label")
	})
}

func TestSetPathNormalizer(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	// Set custom path normalizer
	metrics.SetPathNormalizer(func(c *gin.Context) string {
		path := c.FullPath()
		if path == "" {
			// Custom logic: differentiate 404 from other unmatched
			if c.Writer.Status() == 404 {
				return "not_found"
			}
			return "unmatched"
		}
		return path
	})

	// Create a test router
	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response")
	})

	t.Run("Custom normalizer is used", func(t *testing.T) {
		// Make request to unregistered route (404)
		req, _ := http.NewRequest("GET", "/custom", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		m := &dto.Metric{}
		// Should use custom "not_found" label
		_ = metrics.requestCount.WithLabelValues("GET", "not_found", "404").Write(m)
		assert.Equal(t, 1, int(m.Counter.GetValue()), "Should use custom normalizer")
	})
}
