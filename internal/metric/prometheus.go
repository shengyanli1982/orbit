package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"go.uber.org/zap"
)

// metricLabels 包含了一些默认标签
// metricLabels contains some default labels
var metricLabels = []string{"method", "path", "status"}

type ServerMetrics struct {
	requestCount     *prometheus.CounterVec   // 请求计数器 (request count)
	requestLatencies *prometheus.HistogramVec // 请求延迟直方图 (request latency histogram)
	requestLatency   *prometheus.GaugeVec     // 请求延迟仪表盘 (request latency gauge)
	registry         *prometheus.Registry     // Prometheus注册表 (Prometheus registry)
}

// NewServerMetrics 创建一个新的ServerMetrics实例
// NewServerMetrics creates a new ServerMetrics instance.
func NewServerMetrics(registry *prometheus.Registry) *ServerMetrics {
	return &ServerMetrics{
		// 创建新的计数器 (Create new counter)
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_count", // HTTP请求总数 (Total number of HTTP requests made)
				Help:      "Total number of HTTP requests made.",
			},
			metricLabels,
		),

		// 创建新的直方图 (Create new histogram)
		requestLatencies: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds_histogram", // HTTP请求延迟直方图（毫秒） (HTTP request latency histogram in seconds)
				Help:      "HTTP request latencies in seconds(Histogram).",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			metricLabels,
		),

		// 创建新的仪表盘 (Create new gauge)
		requestLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds", // HTTP请求延迟仪表盘（毫秒） (HTTP request latency gauge in seconds)
				Help:      "HTTP request latencies in seconds.",
			},
			metricLabels,
		),

		// Prometheus注册表 (Prometheus registry)
		registry: registry,
	}
}

// Reset 重置所有的指标
// Register registers the metrics to the Prometheus registry.
func (m *ServerMetrics) Register() {
	m.registry.MustRegister(m.requestCount)
	m.registry.MustRegister(m.requestLatencies)
	m.registry.MustRegister(m.requestLatency)
}

// Unregister 从Prometheus注册表中注销指标
// Unregister unregisters the metrics from the Prometheus registry.
func (m *ServerMetrics) Unregister() {
	m.registry.Unregister(m.requestCount)
	m.registry.Unregister(m.requestLatencies)
	m.registry.Unregister(m.requestLatency)
}

// IncRequestCount 增加请求计数
// IncRequestCount increments the request count.
func (m *ServerMetrics) IncRequestCount(method, path, status string) {
	m.requestCount.WithLabelValues(method, path, status).Inc()
}

// ObserveRequestLatency 观察请求延迟
// ObserveRequestLatency observes the request latency.
func (m *ServerMetrics) ObserveRequestLatency(method, path, status string, latency float64) {
	m.requestLatencies.WithLabelValues(method, path, status).Observe(latency)
}

// SetRequestLatency 设置请求延迟
// SetRequestLatency sets the request latency.
func (m *ServerMetrics) SetRequestLatency(method, path, status string, latency float64) {
	m.requestLatency.WithLabelValues(method, path, status).Set(latency)
}

// ResetRequestCount 重置请求计数
// ResetRequestCount resets the request latency
func (m *ServerMetrics) ResetRequestLatency(method, path, status string) {
	m.requestLatency.DeleteLabelValues(method, path, status)
}

// ResetRequestCount 重置请求计数
// ResetRequestCount resets tobserves the request latency.
func (m *ServerMetrics) ResetRequestLatencies(method, path, status string) {
	m.requestLatencies.DeleteLabelValues(method, path, status)
}

// ResetRequestCount 重置请求计数
// ResetRequestCount resets the request count.
func (m *ServerMetrics) ResetRequestCount(method, path, status string) {
	m.requestCount.DeleteLabelValues(method, path, status)
}

// Reset 重置所有的指标
// Reset resets the metrics.
func (m *ServerMetrics) Reset() {
	m.requestCount.Reset()
	m.requestLatencies.Reset()
	m.requestLatency.Reset()
}

// HandlerFunc 返回一个Gin中间件处理程序
// HandlerFunc returns a Gin middleware handler function.
func (m *ServerMetrics) HandlerFunc(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		// 执行下一个 middleware
		// Execute the next middleware
		context.Next()

		// 处理 response
		// Handle response
		if len(context.Errors) > 0 {
			// 记录错误
			// Log error
			for _, err := range context.Errors {
				logger.Error(err)
			}
		} else {
			// Response 响应延迟
			// Response latency
			latency := time.Since(start).Seconds()
			status := strconv.Itoa(context.Writer.Status())

			// 记录指标
			// Record metrics
			m.IncRequestCount(context.Request.Method, context.Request.URL.Path, status)                // 记录请求计数器 (Record request counter)
			m.ObserveRequestLatency(context.Request.Method, context.Request.URL.Path, status, latency) // 记录请求延迟直方图 (Record request latency histogram)
			m.SetRequestLatency(context.Request.Method, context.Request.URL.Path, status, latency)     // 记录请求延迟仪表盘 (Record request latency gauge)
		}
	}
}
