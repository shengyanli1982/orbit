package metric

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/middleware"
)

// 度量标准的标签
var metricLabels = []string{"method", "path", "status"}

// defaultRequestDurationBuckets 默认 HTTP 请求耗时桶（秒）
// 设计原则：
// 1) 常见时延区间（5ms~1s）更细粒度，便于定位性能回退
// 2) 慢请求长尾区间扩展到 120s，覆盖批处理/下游抖动等场景
// 3) 控制桶数量，避免时序成本过高
var defaultRequestDurationBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1,
	0.25, 0.5, 1, 2.5, 5,
	10, 15, 20, 30, 45, 60, 90, 120,
}

// ServerMetrics 结构体包含了请求计数器、请求延迟直方图、请求延迟仪表盘和 Prometheus 注册表
type ServerMetrics struct {
	requestCount     *prometheus.CounterVec   // 请求计数器
	requestLatencies *prometheus.HistogramVec // 请求延迟直方图
	requestLatency   *prometheus.GaugeVec     // 请求延迟仪表盘
	registry         *prometheus.Registry     // Prometheus注册表
	pathNormalizer   atomic.Value             // 存储 func(*gin.Context) string
	ownedCount       bool                     // requestCount 是否由本实例新注册，共享注册表时仅自有项可注销
	ownedLatencies   bool                     // requestLatencies 是否由本实例新注册
	ownedLatency     bool                     // requestLatency 是否由本实例新注册
	cache            sync.Map
}

type metricCacheKey struct {
	method string
	path   string
	status string
}

type cachedMetric struct {
	counter  prometheus.Counter
	observer prometheus.Observer
	gauge    prometheus.Gauge
}

// defaultPathNormalizer 是默认的路径规范化函数
// 使用 Gin 的 FullPath() 获取路由模板，防止 cardinality explosion
func defaultPathNormalizer(c *gin.Context) string {
	path := c.FullPath()
	if path == "" {
		// 对于未注册路由，使用统一标记防止无限标签组合
		return "unmatched"
	}
	return path
}

// 返回一个新的 ServerMetrics 实例
func NewServerMetrics(registry *prometheus.Registry) *ServerMetrics {
	metrics := &ServerMetrics{
		// 创建一个新的 Prometheus 计数器向量，用于记录 HTTP 请求总数
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: com.OrbitName,
				Name:      "http_requests_total", // HTTP请求总数（Prometheus Counter 命名规范）
				Help:      "Total number of HTTP requests made.",
			},
			metricLabels,
		),

		// 创建一个新的 Prometheus 直方图向量，用于记录 HTTP 请求延迟
		requestLatencies: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_duration_seconds", // HTTP请求耗时分布（秒）
				Help:      "HTTP request duration in seconds (histogram).",
				Buckets:   defaultRequestDurationBuckets,
			},
			metricLabels,
		),

		// 创建一个新的 Prometheus 仪表盘向量，用于记录 HTTP 请求延迟
		requestLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_duration_seconds_last", // 最近一次请求耗时（秒）
				Help:      "Last observed HTTP request duration in seconds.",
			},
			metricLabels,
		),

		// Prometheus 注册表用于注册和收集度量标准
		registry: registry,
	}
	metrics.pathNormalizer.Store(defaultPathNormalizer)
	return metrics
}

// formatStatusCode 将 HTTP 状态码格式化为字符串
// 对于常见状态码使用预定义的字符串，避免重复转换
func formatStatusCode(status int) string {
	switch status {
	case http.StatusOK:
		return "200"
	case http.StatusCreated:
		return "201"
	case http.StatusNoContent:
		return "204"
	case http.StatusBadRequest:
		return "400"
	case http.StatusUnauthorized:
		return "401"
	case http.StatusForbidden:
		return "403"
	case http.StatusNotFound:
		return "404"
	case http.StatusMethodNotAllowed:
		return "405"
	case http.StatusInternalServerError:
		return "500"
	case http.StatusBadGateway:
		return "502"
	case http.StatusServiceUnavailable:
		return "503"
	case http.StatusGatewayTimeout:
		return "504"
	default:
		return strconv.Itoa(status)
	}
}

// registerErrorLog 用于 Register 阶段的非致命告警，复用包内 ErrorLog 机制与默认 logger
var registerErrorLog = NewErrorLog(&com.DefaultLogrLogger)

// registerCollector 注册单个 collector，返回实际生效的 collector 与所有权标志。
// 共享注册表下同一 collector 已被注册时，按 prometheus 标准做法复用既有 collector，
// 使本实例的写入反映到共享抓取输出；复用失败（类型不匹配）或其他注册错误时
// 记录 warning 并返回原 collector（其写入不会出现在抓取中，但不中断服务启动）。
func registerCollector[T prometheus.Collector](registry *prometheus.Registry, collector T) (T, bool) {
	err := registry.Register(collector)
	if err == nil {
		return collector, true
	}
	var alreadyRegistered prometheus.AlreadyRegisteredError
	if errors.As(err, &alreadyRegistered) {
		if existing, ok := alreadyRegistered.ExistingCollector.(T); ok {
			return existing, false
		}
	}
	registerErrorLog.Println("register collector failed:", err)
	return collector, false
}

// Register 将度量标准注册到 Prometheus 注册表。
// 每个 collector 独立记录所有权：共享注册表中已注册的复用既有 collector，
// 新注册的记为自有；已自有项跳过重复注册，避免所有权被误降级。
func (m *ServerMetrics) Register() {
	if !m.ownedCount {
		m.requestCount, m.ownedCount = registerCollector(m.registry, m.requestCount) // 请求计数器
	}
	if !m.ownedLatencies {
		m.requestLatencies, m.ownedLatencies = registerCollector(m.registry, m.requestLatencies) // 请求延迟直方图
	}
	if !m.ownedLatency {
		m.requestLatency, m.ownedLatency = registerCollector(m.registry, m.requestLatency) // 请求延迟仪表盘
	}
}

// Unregister 将度量标准从 Prometheus 注册表中注销。
// 仅注销由本实例新注册的 collector；共享注册表中复用的项不受影响，避免误删其他实例的指标。
func (m *ServerMetrics) Unregister() {
	if m.ownedCount {
		m.registry.Unregister(m.requestCount) // 注销请求计数器
	}
	if m.ownedLatencies {
		m.registry.Unregister(m.requestLatencies) // 注销请求延迟直方图
	}
	if m.ownedLatency {
		m.registry.Unregister(m.requestLatency) // 注销请求延迟仪表盘
	}
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
	// 同步失效标签缓存，避免中间件继续命中已脱离 vector 的孤儿子对象
	m.cache.Delete(metricCacheKey{method: method, path: path, status: status})
}

// 重置请求延迟直方图
func (m *ServerMetrics) ResetRequestLatencies(method, path, status string) {
	m.requestLatencies.DeleteLabelValues(method, path, status) // 删除指定标签值的请求延迟直方图
	// 同步失效标签缓存，避免中间件继续命中已脱离 vector 的孤儿子对象
	m.cache.Delete(metricCacheKey{method: method, path: path, status: status})
}

// 重置请求计数器
func (m *ServerMetrics) ResetRequestCount(method, path, status string) {
	m.requestCount.DeleteLabelValues(method, path, status) // 删除指定标签值的请求计数器
	// 同步失效标签缓存，避免中间件继续命中已脱离 vector 的孤儿子对象
	m.cache.Delete(metricCacheKey{method: method, path: path, status: status})
}

// 重置所有度量标准
func (m *ServerMetrics) Reset() {
	m.requestCount.Reset()     // 重置请求计数器
	m.requestLatencies.Reset() // 重置请求延迟直方图
	m.requestLatency.Reset()   // 重置请求延迟仪表盘
	// 清空标签缓存，避免中间件继续命中已脱离 vector 的孤儿子对象
	m.cache.Range(func(key, _ any) bool {
		m.cache.Delete(key)
		return true
	})
}

// SetPathNormalizer 设置自定义的路径规范化函数（包内部扩展点）。
// 注意：本包为 internal 包，且 Engine 未暴露 metric 实例，
// 框架外部调用方当前无法触达本方法，仅为框架内部预留的扩展点。
// 规范化函数的能力示例：
// - 根据状态码分类未匹配路由（404 -> "not_found"）
// - 使用正则表达式进一步规范化路径
// - 实现自定义的路由参数替换逻辑
func (m *ServerMetrics) SetPathNormalizer(fn func(*gin.Context) string) {
	if fn == nil {
		fn = defaultPathNormalizer
	}
	m.pathNormalizer.Store(fn)
}

// 返回一个 Gin 中间件处理函数
func (m *ServerMetrics) HandlerFunc(logger *logr.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 快速路径：如果是需要跳过的资源，立即返回
		if middleware.SkipResources(context) {
			context.Next()
			return
		}

		// 无锁读取路径归一化函数，避免热路径锁竞争。
		path := m.pathNormalizer.Load().(func(*gin.Context) string)(context)

		method := context.Request.Method
		start := time.Now()
		context.Next()

		// context.Errors 的错误日志由 AccessLogger 统一记录（缺陷 #8：指标采集不承担
		// 日志职责，避免默认启用 metric 时每条请求错误被双重打日志）；
		// logger 参数保留以维持 HandlerFunc 签名稳定

		// 获取状态码（常见状态码走无分配快路径）
		status := formatStatusCode(context.Writer.Status())

		latency := time.Since(start).Seconds()
		key := metricCacheKey{method: method, path: path, status: status}
		cached, ok := m.cache.Load(key)
		if !ok {
			cached = &cachedMetric{
				counter:  m.requestCount.WithLabelValues(method, path, status),
				observer: m.requestLatencies.WithLabelValues(method, path, status),
				gauge:    m.requestLatency.WithLabelValues(method, path, status),
			}
			m.cache.Store(key, cached)
		}
		cm := cached.(*cachedMetric)
		cm.counter.Inc()
		cm.observer.Observe(latency)
		cm.gauge.Set(latency)
	}
}
