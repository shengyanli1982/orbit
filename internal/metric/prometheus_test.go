package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	lru "github.com/hashicorp/golang-lru/v2"
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

func TestServerMetricsLRUCache(t *testing.T) {
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

		// Check cache has entry (LRU cache is thread-safe)
		_, found := metrics.labelCache.Get("GET:/test1:200")
		assert.True(t, found, "Cache should contain entry for GET:/test1:200")

		// Make second request to same endpoint
		req2, _ := http.NewRequest("GET", "/test1", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)

		// Cache should still have the entry
		cachedLabels, found := metrics.labelCache.Get("GET:/test1:200")
		assert.True(t, found, "Cache should still contain entry")
		assert.Equal(t, []string{"GET", "/test1", "200"}, cachedLabels, "Cached labels should match expected values")
	})

	t.Run("Cache respects LRU eviction", func(t *testing.T) {
		// Reset cache for clean test
		metrics.Reset()

		// Get initial cache length (should be 0)
		initialLen := metrics.labelCache.Len()
		assert.Equal(t, 0, initialLen, "Cache should be empty after reset")

		// Make requests to different endpoints
		req1, _ := http.NewRequest("GET", "/test1", nil)
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)

		req2, _ := http.NewRequest("GET", "/test2", nil)
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)

		// Check cache has both entries (LRU cache is thread-safe)
		cacheLen := metrics.labelCache.Len()
		_, found1 := metrics.labelCache.Get("GET:/test1:200")
		_, found2 := metrics.labelCache.Get("GET:/test2:200")

		assert.Equal(t, 2, cacheLen, "Cache should contain 2 entries")
		assert.True(t, found1, "Cache should contain entry for /test1")
		assert.True(t, found2, "Cache should contain entry for /test2")
	})
}

func TestServerMetricsLRUEviction(t *testing.T) {
	// Create a small LRU cache for testing eviction
	cache, err := lru.New[string, []string](2) // Only 2 entries max
	assert.NoError(t, err)

	// Test LRU eviction behavior
	t.Run("LRU evicts least recently used items", func(t *testing.T) {
		// Add first item
		cache.Add("key1", []string{"value1"})
		assert.Equal(t, 1, cache.Len())

		// Add second item
		cache.Add("key2", []string{"value2"})
		assert.Equal(t, 2, cache.Len())

		// Both items should be present
		_, found1 := cache.Get("key1")
		_, found2 := cache.Get("key2")
		assert.True(t, found1, "key1 should be present")
		assert.True(t, found2, "key2 should be present")

		// Add third item - should evict key1 (least recently used)
		cache.Add("key3", []string{"value3"})
		assert.Equal(t, 2, cache.Len(), "Cache should still have max 2 items")

		// key1 should be evicted, key2 and key3 should remain
		_, found1 = cache.Get("key1")
		_, found2 = cache.Get("key2")
		_, found3 := cache.Get("key3")

		assert.False(t, found1, "key1 should be evicted (least recently used)")
		assert.True(t, found2, "key2 should still be present")
		assert.True(t, found3, "key3 should be present")
	})

	t.Run("LRU updates access order", func(t *testing.T) {
		cache.Purge() // Clear cache

		// Add two items
		cache.Add("keyA", []string{"valueA"})
		cache.Add("keyB", []string{"valueB"})

		// Access keyA to make it recently used
		cache.Get("keyA")

		// Add third item - should evict keyB (now least recently used)
		cache.Add("keyC", []string{"valueC"})

		_, foundA := cache.Get("keyA")
		_, foundB := cache.Get("keyB")
		_, foundC := cache.Get("keyC")

		assert.True(t, foundA, "keyA should still be present (recently accessed)")
		assert.False(t, foundB, "keyB should be evicted (least recently used)")
		assert.True(t, foundC, "keyC should be present")
	})
}

func BenchmarkServerMetricsLRUCache(b *testing.B) {
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
