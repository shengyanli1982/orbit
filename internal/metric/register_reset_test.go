package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// findRequestCounterValue 在注册表抓取输出中查找指定标签的计数器值。
// 与 gatherRequestCounterValue 不同，未找到时返回 (0, false) 而不使测试失败，
// 用于断言"序列不存在"的场景。
func findRequestCounterValue(registry *prometheus.Registry, method, path, status string) (float64, bool) {
	families, err := registry.Gather()
	if err != nil {
		return 0, false
	}
	for _, family := range families {
		if family.GetName() != "orbit_http_requests_total" {
			continue
		}
		for _, pm := range family.GetMetric() {
			var gotMethod, gotPath, gotStatus string
			for _, lp := range pm.GetLabel() {
				switch lp.GetName() {
				case "method":
					gotMethod = lp.GetValue()
				case "path":
					gotPath = lp.GetValue()
				case "status":
					gotStatus = lp.GetValue()
				}
			}
			if gotMethod == method && gotPath == path && gotStatus == status {
				return pm.GetCounter().GetValue(), true
			}
		}
	}
	return 0, false
}

// gatherHasMetricWithLabels 判断抓取输出中指定指标族是否存在给定标签组合的序列。
func gatherHasMetricWithLabels(registry *prometheus.Registry, familyName, method, path, status string) bool {
	families, err := registry.Gather()
	if err != nil {
		return false
	}
	for _, family := range families {
		if family.GetName() != familyName {
			continue
		}
		for _, pm := range family.GetMetric() {
			var gotMethod, gotPath, gotStatus string
			for _, lp := range pm.GetLabel() {
				switch lp.GetName() {
				case "method":
					gotMethod = lp.GetValue()
				case "path":
					gotPath = lp.GetValue()
				case "status":
					gotStatus = lp.GetValue()
				}
			}
			if gotMethod == method && gotPath == path && gotStatus == status {
				return true
			}
		}
	}
	return false
}

// TestServerMetricsSharedRegistryReusesExistingCollectors 验证缺陷 #5 的核心修复：
// 共享注册表下第二个实例 Register 后必须复用既有 collector，
// 其写入反映到抓取输出，而不是静默丢失到自身未注册的 vector。
func TestServerMetricsSharedRegistryReusesExistingCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()

	first := NewServerMetrics(registry)
	first.Register()
	defer first.Unregister()

	second := NewServerMetrics(registry)
	require.NotPanics(t, func() { second.Register() })

	// 第二个实例的写入必须反映到共享注册表（缺陷行为：写入自身未注册 vector，/metrics 永远看不到）
	second.IncRequestCount("GET", "/second", "200")
	second.ObserveRequestLatency("GET", "/second", "200", 0.5)
	second.SetRequestLatency("GET", "/second", "200", 0.5)

	value, found := findRequestCounterValue(registry, "GET", "/second", "200")
	require.True(t, found, "second instance counter write must appear in shared registry")
	assert.Equal(t, float64(1), value)
	assert.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds", "GET", "/second", "200"),
		"second instance histogram observation must appear in shared registry")
	assert.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds_last", "GET", "/second", "200"),
		"second instance gauge write must appear in shared registry")

	// second 未新注册任何 collector：Unregister 不得影响复用的共享 collector
	second.Unregister()
	first.IncRequestCount("GET", "/first", "200")
	firstValue, firstFound := findRequestCounterValue(registry, "GET", "/first", "200")
	require.True(t, firstFound, "first instance collector must survive second.Unregister")
	assert.Equal(t, float64(1), firstValue)
	secondValue, secondFound := findRequestCounterValue(registry, "GET", "/second", "200")
	require.True(t, secondFound, "reused collector data must survive second.Unregister")
	assert.Equal(t, float64(1), secondValue)
}

// TestServerMetricsSharedRegistryHandlerFuncReflected 验证共享注册表下
// 第二个实例的中间件路径（含 sync.Map 缓存）写入同样反映到共享 collector。
func TestServerMetricsSharedRegistryHandlerFuncReflected(t *testing.T) {
	registry := prometheus.NewRegistry()

	first := NewServerMetrics(registry)
	first.Register()
	defer first.Unregister()

	second := NewServerMetrics(registry)
	second.Register()
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(second.HandlerFunc(&logger))
	router.GET("/shared", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/shared", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	value, found := findRequestCounterValue(registry, "GET", "/shared", "200")
	require.True(t, found, "handler of second instance must write to shared collector")
	assert.Equal(t, float64(1), value)
}

// TestServerMetricsMixedOwnershipUnregisterOnlyOwned 验证缺陷 #5 的 per-collector 所有权修复：
// 混合状态（部分新注册 + 部分复用）下 Unregister 只注销自有项，
// 既不泄漏自有 collector，也不误删复用项。
func TestServerMetricsMixedOwnershipUnregisterOnlyOwned(t *testing.T) {
	registry := prometheus.NewRegistry()

	first := NewServerMetrics(registry)
	first.Register()
	defer first.Unregister()
	first.IncRequestCount("GET", "/a", "200")

	// 制造混合状态：gauge collector 从注册表移除，
	// second.Register 时 counter/histogram 走复用（非自有），gauge 走全新注册（自有）
	registry.Unregister(first.requestLatency)

	second := NewServerMetrics(registry)
	require.NotPanics(t, func() { second.Register() })
	second.IncRequestCount("GET", "/b", "200")
	second.ObserveRequestLatency("GET", "/b", "200", 0.1)
	second.SetRequestLatency("GET", "/b", "200", 0.2)

	value, found := findRequestCounterValue(registry, "GET", "/b", "200")
	require.True(t, found, "reused counter must reflect second instance write")
	assert.Equal(t, float64(1), value)
	require.True(t, registryHasMetricFamily(registry, "orbit_http_request_duration_seconds_last"),
		"freshly registered gauge must be visible")

	// second.Unregister 只注销自有 gauge；复用的 counter/histogram 不受影响
	second.Unregister()
	assert.False(t, registryHasMetricFamily(registry, "orbit_http_request_duration_seconds_last"),
		"owned gauge must be unregistered by second.Unregister")
	assert.True(t, registryHasMetricFamily(registry, "orbit_http_requests_total"),
		"reused counter must survive second.Unregister")
	assert.True(t, registryHasMetricFamily(registry, "orbit_http_request_duration_seconds"),
		"reused histogram must survive second.Unregister")

	// first 继续使用共享 counter 正常
	first.IncRequestCount("GET", "/a", "200")
	firstValue, firstFound := findRequestCounterValue(registry, "GET", "/a", "200")
	require.True(t, firstFound)
	assert.Equal(t, float64(2), firstValue)
}

// newCacheTestEnv 构造一个已注册并挂载中间件的测试环境，返回 serve 函数。
func newCacheTestEnv(t *testing.T) (*prometheus.Registry, *ServerMetrics, func()) {
	t.Helper()
	registry := prometheus.NewRegistry()
	metrics := NewServerMetrics(registry)
	metrics.Register()
	t.Cleanup(metrics.Unregister)
	logger := zapr.NewLogger(zap.NewExample())

	router := gin.New()
	router.Use(metrics.HandlerFunc(&logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	serve := func() {
		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
	}
	return registry, metrics, serve
}

// TestServerMetricsResetInvalidatesCache 验证缺陷 #7 的全量 Reset 修复：
// Reset 后缓存必须失效，再次请求时指标重新出现在抓取输出。
func TestServerMetricsResetInvalidatesCache(t *testing.T) {
	registry, metrics, serve := newCacheTestEnv(t)

	serve()
	value, found := findRequestCounterValue(registry, "GET", "/test", "200")
	require.True(t, found)
	require.Equal(t, float64(1), value)

	metrics.Reset()

	_, found = findRequestCounterValue(registry, "GET", "/test", "200")
	require.False(t, found, "series must disappear from gather after Reset")

	// 缺陷行为：缓存仍持有已脱离 vector 的孤儿子对象，
	// 再次请求命中缓存写到孤儿上，序列永远不会重新出现
	serve()
	value, found = findRequestCounterValue(registry, "GET", "/test", "200")
	require.True(t, found, "series must reappear after Reset + new request")
	assert.Equal(t, float64(1), value)
	assert.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds", "GET", "/test", "200"),
		"histogram series must reappear after Reset + new request")
	assert.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds_last", "GET", "/test", "200"),
		"gauge series must reappear after Reset + new request")
}

// TestServerMetricsResetRequestCountInvalidatesCacheEntry 验证缺陷 #7 的精确删除修复：
// ResetRequestCount 删除单个标签组合后，对应缓存键必须失效。
func TestServerMetricsResetRequestCountInvalidatesCacheEntry(t *testing.T) {
	registry, metrics, serve := newCacheTestEnv(t)

	serve()
	value, found := findRequestCounterValue(registry, "GET", "/test", "200")
	require.True(t, found)
	require.Equal(t, float64(1), value)

	metrics.ResetRequestCount("GET", "/test", "200")

	_, found = findRequestCounterValue(registry, "GET", "/test", "200")
	require.False(t, found, "counter series must disappear after ResetRequestCount")

	// 缺陷行为：缓存仍持有孤儿 counter 子对象，序列不会重新出现
	serve()
	value, found = findRequestCounterValue(registry, "GET", "/test", "200")
	require.True(t, found, "counter series must reappear after ResetRequestCount + new request")
	assert.Equal(t, float64(1), value)
}

// TestServerMetricsResetRequestLatenciesInvalidatesCacheEntry 验证缺陷 #7：
// ResetRequestLatencies 删除直方图子项后，对应缓存键必须失效。
func TestServerMetricsResetRequestLatenciesInvalidatesCacheEntry(t *testing.T) {
	registry, metrics, serve := newCacheTestEnv(t)

	serve()
	require.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds", "GET", "/test", "200"))

	metrics.ResetRequestLatencies("GET", "/test", "200")
	require.False(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds", "GET", "/test", "200"),
		"histogram series must disappear after ResetRequestLatencies")

	// 缺陷行为：缓存仍持有孤儿 observer，序列不会重新出现
	serve()
	assert.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds", "GET", "/test", "200"),
		"histogram series must reappear after ResetRequestLatencies + new request")
}

// TestServerMetricsResetRequestLatencyInvalidatesCacheEntry 验证缺陷 #7：
// ResetRequestLatency 删除 gauge 子项后，对应缓存键必须失效。
func TestServerMetricsResetRequestLatencyInvalidatesCacheEntry(t *testing.T) {
	registry, metrics, serve := newCacheTestEnv(t)

	serve()
	require.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds_last", "GET", "/test", "200"))

	metrics.ResetRequestLatency("GET", "/test", "200")
	require.False(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds_last", "GET", "/test", "200"),
		"gauge series must disappear after ResetRequestLatency")

	// 缺陷行为：缓存仍持有孤儿 gauge，序列不会重新出现
	serve()
	assert.True(t, gatherHasMetricWithLabels(registry, "orbit_http_request_duration_seconds_last", "GET", "/test", "200"),
		"gauge series must reappear after ResetRequestLatency + new request")
}
