package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/middleware"
)

// metricLabels 包含了度量标准的标签。
// metricLabels contains the labels of the metrics.
var metricLabels = []string{"method", "path", "status"}

// ServerMetrics 结构体包含了请求计数器、请求延迟直方图、请求延迟仪表盘和 Prometheus 注册表。
// The ServerMetrics struct contains a request counter, request latency histogram, request latency gauge, and Prometheus registry.
type ServerMetrics struct {
	requestCount     *prometheus.CounterVec   // 请求计数器 (request count)
	requestLatencies *prometheus.HistogramVec // 请求延迟直方图 (request latency histogram)
	requestLatency   *prometheus.GaugeVec     // 请求延迟仪表盘 (request latency gauge)
	registry         *prometheus.Registry     // Prometheus注册表 (Prometheus registry)
}

// NewServerMetrics 函数返回一个新的 ServerMetrics 实例。
// The NewServerMetrics function returns a new ServerMetrics instance.
func NewServerMetrics(registry *prometheus.Registry) *ServerMetrics {
	return &ServerMetrics{
		// 创建一个新的 Prometheus 计数器向量，用于记录 HTTP 请求总数。
		// Create a new Prometheus counter vector to record the total number of HTTP requests.
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_count", // HTTP请求总数 (Total number of HTTP requests made)
				Help:      "Total number of HTTP requests made.",
			},
			metricLabels,
		),

		// 创建一个新的 Prometheus 直方图向量，用于记录 HTTP 请求延迟。
		// Create a new Prometheus histogram vector to record HTTP request latency.
		requestLatencies: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds_histogram", // HTTP请求延迟直方图（毫秒） (HTTP request latency histogram in seconds)
				Help:      "HTTP request latencies in seconds(Histogram).",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			metricLabels,
		),

		// 创建一个新的 Prometheus 仪表盘向量，用于记录 HTTP 请求延迟。
		// Create a new Prometheus gauge vector to record HTTP request latency.
		requestLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds", // HTTP请求延迟仪表盘（毫秒） (HTTP request latency gauge in seconds)
				Help:      "HTTP request latencies in seconds.",
			},
			metricLabels,
		),

		// Prometheus 注册表用于注册和收集度量标准。
		// The Prometheus registry is used to register and collect metrics.
		registry: registry,
	}
}

// Register 方法将度量标准注册到 Prometheus 注册表。
// The Register method registers the metrics to the Prometheus registry.
func (m *ServerMetrics) Register() {
	m.registry.MustRegister(m.requestCount)     // 注册请求计数器 (Register request counter)
	m.registry.MustRegister(m.requestLatencies) // 注册请求延迟直方图 (Register request latency histogram)
	m.registry.MustRegister(m.requestLatency)   // 注册请求延迟仪表盘 (Register request latency gauge)
}

// Unregister 方法将度量标准从 Prometheus 注册表中注销。
// The Unregister method unregisters the metrics from the Prometheus registry.
func (m *ServerMetrics) Unregister() {
	m.registry.Unregister(m.requestCount)     // 注销请求计数器 (Unregister request counter)
	m.registry.Unregister(m.requestLatencies) // 注销请求延迟直方图 (Unregister request latency histogram)
	m.registry.Unregister(m.requestLatency)   // 注销请求延迟仪表盘 (Unregister request latency gauge)
}

// IncRequestCount 方法增加请求计数。
// The IncRequestCount method increments the request count.
func (m *ServerMetrics) IncRequestCount(method, path, status string) {
	m.requestCount.WithLabelValues(method, path, status).Inc() // 增加请求计数 (Increment request count)
}

// ObserveRequestLatency 方法观察请求延迟。
// The ObserveRequestLatency method observes the request latency.
func (m *ServerMetrics) ObserveRequestLatency(method, path, status string, latency float64) {
	m.requestLatencies.WithLabelValues(method, path, status).Observe(latency) // 观察请求延迟 (Observe request latency)
}

// SetRequestLatency 方法设置请求延迟。
// The SetRequestLatency method sets the request latency.
func (m *ServerMetrics) SetRequestLatency(method, path, status string, latency float64) {
	m.requestLatency.WithLabelValues(method, path, status).Set(latency) // 设置请求延迟 (Set request latency)
}

// ResetRequestLatency 方法重置请求延迟。
// The ResetRequestLatency method resets the request latency.
func (m *ServerMetrics) ResetRequestLatency(method, path, status string) {
	m.requestLatency.DeleteLabelValues(method, path, status) // 删除指定标签值的请求延迟 (Delete request latency for the specified label values)
}

// ResetRequestLatencies 方法重置请求延迟直方图。
// The ResetRequestLatencies method resets the request latency histogram.
func (m *ServerMetrics) ResetRequestLatencies(method, path, status string) {
	m.requestLatencies.DeleteLabelValues(method, path, status) // 删除指定标签值的请求延迟直方图 (Delete request latency histogram for the specified label values)
}

// ResetRequestCount 方法重置请求计数器。
// The ResetRequestCount method resets the request counter.
func (m *ServerMetrics) ResetRequestCount(method, path, status string) {
	m.requestCount.DeleteLabelValues(method, path, status) // 删除指定标签值的请求计数器 (Delete request counter for the specified label values)
}

// Reset 方法重置所有度量标准。
// The Reset method resets all metrics.
func (m *ServerMetrics) Reset() {
	m.requestCount.Reset()     // 重置请求计数器 (Reset request counter)
	m.requestLatencies.Reset() // 重置请求延迟直方图 (Reset request latency histogram)
	m.requestLatency.Reset()   // 重置请求延迟仪表盘 (Reset request latency gauge)
}

// HandlerFunc 返回一个 Gin 中间件处理函数。
// HandlerFunc returns a Gin middleware handler function.
func (m *ServerMetrics) HandlerFunc(logger *logr.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 提前判断是否需要跳过，避免不必要的时间记录
		if middleware.SkipResources(context) {
			context.Next()
			return
		}

		// 记录开始时间
		start := time.Now()
		
		// 提前获取常用值，避免重复获取
		method := context.Request.Method
		path := context.Request.URL.Path

		// 执行后续中间件
		context.Next()

		// 处理错误情况
		if errors := context.Errors; len(errors) > 0 {
			for _, err := range errors {
				logger.Error(err, "Error occurred")
			}
			return
		}

		// 计算指标
		status := strconv.Itoa(context.Writer.Status())
		latency := time.Since(start).Seconds()

		// 批量更新指标，减少函数调用
		m.updateMetrics(method, path, status, latency)
	}
}

// updateMetrics 集中更新所有指标，提高代码复用性
func (m *ServerMetrics) updateMetrics(method, path, status string, latency float64) {
	m.IncRequestCount(method, path, status)
	m.ObserveRequestLatency(method, path, status, latency)
	m.SetRequestLatency(method, path, status, latency)
}
