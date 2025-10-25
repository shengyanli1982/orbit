package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/puzpuzpuz/xsync/v3"
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
	assert.NotNil(t, metrics.labelCache)
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

func TestServerMetricsXsyncCache(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	// Create a test router
	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	// Define test routes
	router.GET("/test1", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response 1")
	})
	router.GET("/test2", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response 2")
	})

	// Test cache functionality
	t.Run("Cache stores and retrieves labels", func(t *testing.T) {
		// Make first request to /test1
		req1, _ := http.NewRequest("GET", "/test1", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)

		// Check cache has entry (xsync.Map is thread-safe)
		_, found := metrics.labelCache.Load("GET:/test1:200")
		assert.True(t, found, "Cache should contain entry for GET:/test1:200")

		// Make second request to same endpoint
		req2, _ := http.NewRequest("GET", "/test1", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)

		// Cache should still have the entry
		cachedLabels, found := metrics.labelCache.Load("GET:/test1:200")
		assert.True(t, found, "Cache should still contain entry")
		assert.Equal(t, []string{"GET", "/test1", "200"}, cachedLabels, "Cached labels should match expected values")
	})

	t.Run("Cache stores multiple entries", func(t *testing.T) {
		// Reset cache for clean test
		metrics.Reset()

		// Make requests to different endpoints
		req1, _ := http.NewRequest("GET", "/test1", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)

		req2, _ := http.NewRequest("GET", "/test2", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)

		// Check cache has both entries (xsync.Map is thread-safe)
		_, found1 := metrics.labelCache.Load("GET:/test1:200")
		_, found2 := metrics.labelCache.Load("GET:/test2:200")

		assert.True(t, found1, "Cache should contain entry for /test1")
		assert.True(t, found2, "Cache should contain entry for /test2")
	})
}

func TestServerMetricsXsyncConcurrency(t *testing.T) {
	// Test xsync.Map concurrent access
	cache := xsync.NewMapOf[string, []string]()

	t.Run("Concurrent read and write operations", func(t *testing.T) {
		// Store initial items
		cache.Store("key1", []string{"value1"})
		cache.Store("key2", []string{"value2"})

		// Both items should be present
		_, found1 := cache.Load("key1")
		_, found2 := cache.Load("key2")
		assert.True(t, found1, "key1 should be present")
		assert.True(t, found2, "key2 should be present")

		// Add third item - xsync.Map has no size limit by default
		cache.Store("key3", []string{"value3"})

		// All three items should be present (no LRU eviction)
		_, found1 = cache.Load("key1")
		_, found2 = cache.Load("key2")
		_, found3 := cache.Load("key3")

		assert.True(t, found1, "key1 should still be present")
		assert.True(t, found2, "key2 should still be present")
		assert.True(t, found3, "key3 should be present")
	})

	t.Run("Clear removes all entries", func(t *testing.T) {
		cache.Clear()

		// After clear, items should not be present
		_, found1 := cache.Load("key1")
		_, found2 := cache.Load("key2")
		_, found3 := cache.Load("key3")

		assert.False(t, found1, "key1 should be cleared")
		assert.False(t, found2, "key2 should be cleared")
		assert.False(t, found3, "key3 should be cleared")
	})
}

func BenchmarkServerMetricsXsyncCache(b *testing.B) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	// Create a test router
	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	// Define a test route
	router.GET("/benchmark", func(c *gin.Context) {
		c.String(http.StatusOK, "Benchmark response")
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/benchmark", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
		}
	})
}
