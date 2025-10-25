package metric

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// BenchmarkConcurrentRequestsLight 模拟轻负载并发场景
func BenchmarkConcurrentRequestsLight(b *testing.B) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	router.POST("/api/v1/users", func(c *gin.Context) {
		c.String(http.StatusCreated, "Created")
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		for pb.Next() {
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
		}
	})
}

// BenchmarkConcurrentRequestsHeavy 模拟高负载多路由并发场景
func BenchmarkConcurrentRequestsHeavy(b *testing.B) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	// 注册多个路由模拟真实场景
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/users"},
		{"POST", "/api/v1/users"},
		{"GET", "/api/v1/users/:id"},
		{"PUT", "/api/v1/users/:id"},
		{"DELETE", "/api/v1/users/:id"},
		{"GET", "/api/v1/orders"},
		{"POST", "/api/v1/orders"},
		{"GET", "/api/v1/products"},
		{"GET", "/health"},
		{"GET", "/metrics"},
	}

	for _, route := range routes {
		method, path := route.method, route.path
		switch method {
		case "GET":
			router.GET(path, func(c *gin.Context) { c.String(http.StatusOK, "OK") })
		case "POST":
			router.POST(path, func(c *gin.Context) { c.String(http.StatusCreated, "Created") })
		case "PUT":
			router.PUT(path, func(c *gin.Context) { c.String(http.StatusOK, "Updated") })
		case "DELETE":
			router.DELETE(path, func(c *gin.Context) { c.String(http.StatusNoContent, "") })
		}
	}

	requests := []*http.Request{
		httptest.NewRequest("GET", "/api/v1/users", nil),
		httptest.NewRequest("POST", "/api/v1/users", nil),
		httptest.NewRequest("GET", "/api/v1/users/123", nil),
		httptest.NewRequest("GET", "/api/v1/orders", nil),
		httptest.NewRequest("GET", "/health", nil),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := requests[i%len(requests)]
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			i++
		}
	})
}

// BenchmarkCacheHitRate 测试缓存命中率对性能的影响
func BenchmarkCacheHitRate(b *testing.B) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 预热缓存
	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/test", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
		}
	})
}

// BenchmarkMemoryAllocation 测试内存分配情况
func BenchmarkMemoryAllocation(b *testing.B) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
}

// TestConcurrentSafety 并发安全性测试
func TestConcurrentSafety(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))

	// 注册多个路由
	for i := 0; i < 20; i++ {
		path := fmt.Sprintf("/api/endpoint_%d", i)
		router.GET(path, func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})
	}

	// 并发测试参数
	const (
		goroutines = 100
		iterations = 1000
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 启动多个goroutine并发访问
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				path := fmt.Sprintf("/api/endpoint_%d", i%20)
				req, _ := http.NewRequest("GET", path, nil)
				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req)
			}
		}(g)
	}

	wg.Wait()

	// 测试已完成,验证没有 panic 即表示并发安全
	t.Log("Concurrent safety test completed successfully")
}
