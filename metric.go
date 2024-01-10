package orbit

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
)

var (
	metricLabels = []string{"method", "path", "status"}
)

type ServerMetrics struct {
	requestCount     *prometheus.CounterVec
	requestLatencies *prometheus.HistogramVec
	requestLatency   *prometheus.GaugeVec
	registry         *prometheus.Registry
}

func NewServerMetrics(registry *prometheus.Registry) *ServerMetrics {
	return &ServerMetrics{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_count",
				Help:      "Total number of HTTP requests made.",
			},
			metricLabels,
		),
		requestLatencies: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency_seconds",
				Help:      "HTTP request latencies in Milliseconds.",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			metricLabels,
		),
		requestLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: com.OrbitName,
				Name:      "http_request_latency",
				Help:      "HTTP request latencies in Milliseconds.",
			},
			metricLabels,
		),
		registry: registry,
	}
}

func (m *ServerMetrics) Register() {
	m.registry.MustRegister(m.requestCount)
	m.registry.MustRegister(m.requestLatencies)
	m.registry.MustRegister(m.requestLatency)
}

func (m *ServerMetrics) Unregister() {
	m.registry.Unregister(m.requestCount)
	m.registry.Unregister(m.requestLatencies)
	m.registry.Unregister(m.requestLatency)
}

func (m *ServerMetrics) IncRequestCount(method, path, status string) {
	m.requestCount.WithLabelValues(method, path, status).Inc()
}

func (m *ServerMetrics) ObserveRequestLatency(method, path, status string, latency float64) {
	m.requestLatencies.WithLabelValues(method, path, status).Observe(latency)
}

func (m *ServerMetrics) SetRequestLatency(method, path, status string, latency float64) {
	m.requestLatency.WithLabelValues(method, path, status).Set(latency)
}

func (m *ServerMetrics) ResetRequestLatency(method, path, status string) {
	m.requestLatency.DeleteLabelValues(method, path, status)
}

func (m *ServerMetrics) ResetRequestLatencies(method, path, status string) {
	m.requestLatencies.DeleteLabelValues(method, path, status)
}

func (m *ServerMetrics) ResetRequestCount(method, path, status string) {
	m.requestCount.DeleteLabelValues(method, path, status)
}

func (m *ServerMetrics) Reset() {
	m.requestCount.Reset()
	m.requestLatencies.Reset()
	m.requestLatency.Reset()
}

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
