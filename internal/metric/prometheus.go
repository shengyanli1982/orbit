package metric

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/middleware"
)

// 度量标准的标签
var metricLabels = []string{"method", "path", "status"}

// 默认的 Prometheus 指标路径
const DefaultMetricsPath = com.PromMetricURLPath

// 默认的 Prometheus 指标端口
const DefaultMetricsPort = "9090"

// ServerMetrics 结构体包含了请求计数器、请求延迟直方图、请求延迟仪表盘和 Prometheus 注册表
type ServerMetrics struct {
	requestCount     *prometheus.CounterVec   // 请求计数器
	requestLatencies *prometheus.HistogramVec // 请求延迟直方图
	requestLatency   *prometheus.GaugeVec     // 请求延迟仪表盘
	registry         *prometheus.Registry     // Prometheus注册表
	labelCache       *sync.Map                // 标签缓存
}

// 返回一个新的 ServerMetrics 实例
func NewServerMetrics(registry *prometheus.Registry) *ServerMetrics {
	return &ServerMetrics{
		// 创建一个新的 Prometheus 计数器向量，用于记录 HTTP 请求总数
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_count", // HTTP请求总数
				Help:      "Total number of HTTP requests made.",
			},
			metricLabels,
		),

		// 创建一个新的 Prometheus 直方图向量，用于记录 HTTP 请求延迟
		requestLatencies: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds_histogram", // HTTP请求延迟直方图（秒）
				Help:      "HTTP request latencies in seconds(Histogram).",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			metricLabels,
		),

		// 创建一个新的 Prometheus 仪表盘向量，用于记录 HTTP 请求延迟
		requestLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds", // HTTP请求延迟仪表盘（秒）
				Help:      "HTTP request latencies in seconds.",
			},
			metricLabels,
		),

		// Prometheus 注册表用于注册和收集度量标准
		registry: registry,

		// 初始化标签缓存
		labelCache: &sync.Map{},
	}
}

// 将度量标准注册到 Prometheus 注册表
func (m *ServerMetrics) Register() {
	m.registry.MustRegister(m.requestCount)     // 注册请求计数器
	m.registry.MustRegister(m.requestLatencies) // 注册请求延迟直方图
	m.registry.MustRegister(m.requestLatency)   // 注册请求延迟仪表盘
}

// 将度量标准从 Prometheus 注册表中注销
func (m *ServerMetrics) Unregister() {
	m.registry.Unregister(m.requestCount)     // 注销请求计数器
	m.registry.Unregister(m.requestLatencies) // 注销请求延迟直方图
	m.registry.Unregister(m.requestLatency)   // 注销请求延迟仪表盘
}

// 增加请求计数
func (m *ServerMetrics) IncRequestCount(method, path, status string) {
	m.requestCount.WithLabelValues(method, path, status).Inc() // 增加请求计数
}

// 观察请求延迟
func (m *ServerMetrics) ObserveRequestLatency(method, path, status string, latency float64) {
	m.requestLatencies.WithLabelValues(method, path, status).Observe(latency) // 观察请求延迟
}

// 设置请求延迟
func (m *ServerMetrics) SetRequestLatency(method, path, status string, latency float64) {
	m.requestLatency.WithLabelValues(method, path, status).Set(latency) // 设置请求延迟
}

// 重置请求延迟
func (m *ServerMetrics) ResetRequestLatency(method, path, status string) {
	m.requestLatency.DeleteLabelValues(method, path, status) // 删除指定标签值的请求延迟
}

// 重置请求延迟直方图
func (m *ServerMetrics) ResetRequestLatencies(method, path, status string) {
	m.requestLatencies.DeleteLabelValues(method, path, status) // 删除指定标签值的请求延迟直方图
}

// 重置请求计数器
func (m *ServerMetrics) ResetRequestCount(method, path, status string) {
	m.requestCount.DeleteLabelValues(method, path, status) // 删除指定标签值的请求计数器
}

// 重置所有度量标准
func (m *ServerMetrics) Reset() {
	m.requestCount.Reset()     // 重置请求计数器
	m.requestLatencies.Reset() // 重置请求延迟直方图
	m.requestLatency.Reset()   // 重置请求延迟仪表盘
	m.labelCache = &sync.Map{} // 重置标签缓存
}

// 返回一个 Gin 中间件处理函数
func (m *ServerMetrics) HandlerFunc(logger *logr.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 快速路径：如果是需要跳过的资源，立即返回
		if middleware.SkipResources(context) {
			context.Next()
			return
		}

		// 在进入处理逻辑前获取所有需要的值
		method := context.Request.Method
		// 优先使用 FullPath，它返回路由模式而不是具体的 URL
		path := context.FullPath()
		if path == "" {
			path = context.Request.URL.Path
		}

		// 生成缓存键
		cacheKey := method + ":" + path

		start := time.Now()
		context.Next()

		// 使用 len 替代 errors := context.Errors 的额外分配
		if len(context.Errors) > 0 {
			// 直接遍历 context.Errors，避免中间变量
			for _, err := range context.Errors {
				logger.Error(err, "Error occurred")
			}
			return
		}

		// 获取状态码
		status := strconv.Itoa(context.Writer.Status())

		// 尝试从缓存获取标签值
		var labels []string
		cacheKeyWithStatus := cacheKey + ":" + status
		if cachedLabels, ok := m.labelCache.Load(cacheKeyWithStatus); ok {
			labels = cachedLabels.([]string)
		} else {
			// 创建新的标签值并缓存
			labels = []string{method, path, status}
			m.labelCache.Store(cacheKeyWithStatus, labels)
		}

		// 一次性计算所有指标需要的值
		latency := time.Since(start).Seconds()

		// 使用一次 WithLabelValues 调用更新所有指标
		m.requestCount.WithLabelValues(labels...).Inc()
		m.requestLatencies.WithLabelValues(labels...).Observe(latency)
		m.requestLatency.WithLabelValues(labels...).Set(latency)
	}
}
