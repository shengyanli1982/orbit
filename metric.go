package orbit

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
)

// metricLabels contains the labels of the metrics.
var metricLabels = []string{"method", "path", "status"}

type ServerMetrics struct {
	requestCount     *prometheus.CounterVec   // 请求计数器 (request count)
	requestLatencies *prometheus.HistogramVec // 请求延迟直方图 (request latency histogram)
	requestLatency   *prometheus.GaugeVec     // 请求延迟仪表盘 (request latency gauge)
	registry         *prometheus.Registry     // Prometheus注册表 (Prometheus registry)
}

func NewServerMetrics(registry *prometheus.Registry) *ServerMetrics {
	return &ServerMetrics{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_count", // HTTP请求总数 (Total number of HTTP requests made)
				Help:      "Total number of HTTP requests made.",
			},
			metricLabels,
		),
		requestLatencies: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_milliseconds", // HTTP请求延迟直方图（毫秒） (HTTP request latency histogram in Milliseconds)
				Help:      "HTTP request latencies in Milliseconds.",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			metricLabels,
		),
		requestLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency", // HTTP请求延迟仪表盘（毫秒） (HTTP request latency gauge in Milliseconds)
				Help:      "HTTP request latencies in Milliseconds.",
			},
			metricLabels,
		),
		registry: registry,
	}
}

// Register registers the metrics to the Prometheus registry.
func (m *ServerMetrics) Register() {
	m.registry.MustRegister(m.requestCount)
	m.registry.MustRegister(m.requestLatencies)
	m.registry.MustRegister(m.requestLatency)
}

// Unregister unregisters the metrics from the Prometheus registry.
func (m *ServerMetrics) Unregister() {
	m.registry.Unregister(m.requestCount)
	m.registry.Unregister(m.requestLatencies)
	m.registry.Unregister(m.requestLatency)
}

// IncRequestCount increments the request count.
func (m *ServerMetrics) IncRequestCount(method, path, status string) {
	m.requestCount.WithLabelValues(method, path, status).Inc()
}

// ObserveRequestLatency observes the request latency.
func (m *ServerMetrics) ObserveRequestLatency(method, path, status string, latency float64) {
	m.requestLatencies.WithLabelValues(method, path, status).Observe(latency)
}

// SetRequestLatency sets the request latency.
func (m *ServerMetrics) SetRequestLatency(method, path, status string, latency float64) {
	m.requestLatency.WithLabelValues(method, path, status).Set(latency)
}

// ResetRequestCount resets the request latency
func (m *ServerMetrics) ResetRequestLatency(method, path, status string) {
	m.requestLatency.DeleteLabelValues(method, path, status)
}

// ResetRequestCount resets tobserves the request latency.
func (m *ServerMetrics) ResetRequestLatencies(method, path, status string) {
	m.requestLatencies.DeleteLabelValues(method, path, status)
}

// ResetRequestCount resets the request count.
func (m *ServerMetrics) ResetRequestCount(method, path, status string) {
	m.requestCount.DeleteLabelValues(method, path, status)
}

// Reset resets the metrics.
func (m *ServerMetrics) Reset() {
	m.requestCount.Reset()
	m.requestLatencies.Reset()
	m.requestLatency.Reset()
}

// HandlerFunc returns a Gin middleware handler function.
func (m *ServerMetrics) HandlerFunc() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Start time
		start := time.Now()

		// To next middleware
		context.Next()

		// Handle response
		if len(context.Errors) > 0 {
			// Log error
			for _, err := range context.Errors {
				com.DefaultSugeredLogger.Errorf("server metrics error: %s", err.Error())
			}
		} else {
			// Response latency
			latency := float64(time.Since(start).Milliseconds())
			status := strconv.Itoa(context.Writer.Status())

			// Record metrics
			m.IncRequestCount(context.Request.Method, context.Request.URL.Path, status)
			m.ObserveRequestLatency(context.Request.Method, context.Request.URL.Path, status, latency)
			m.SetRequestLatency(context.Request.Method, context.Request.URL.Path, status, latency)
		}
	}
}
