package orbit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	uhttptool "github.com/shengyanli1982/orbit/utils/httptool"
	ulog "github.com/shengyanli1982/orbit/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigDefaultProxySettings(t *testing.T) {
	config := NewConfig()
	assert.Equal(t, []string{"0.0.0.0/0", "::/0"}, config.TrustedProxies)
	assert.Equal(t, []string{"X-Forwarded-For", "X-Real-IP"}, config.RemoteIPHeaders)
	assert.NotNil(t, config.CORSPolicy)
	assert.True(t, config.CORSPolicy.Enabled)
	assert.True(t, config.CORSPolicy.AllowAllOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, config.CORSPolicy.AllowedMethods)
}

func TestConfigWithProxySettingsCloneInput(t *testing.T) {
	trusted := []string{"10.0.0.0/8"}
	headers := []string{"X-Real-IP"}

	config := NewConfig().
		WithTrustedProxies(trusted).
		WithRemoteIPHeaders(headers)

	trusted[0] = "192.168.0.0/16"
	headers[0] = "X-Forwarded-For"

	assert.Equal(t, []string{"10.0.0.0/8"}, config.TrustedProxies)
	assert.Equal(t, []string{"X-Real-IP"}, config.RemoteIPHeaders)
}

func TestConfigValidationRespectsEmptyTrustedProxies(t *testing.T) {
	config := isConfigValid(&Config{
		Address:         "127.0.0.1",
		Port:            8080,
		TrustedProxies:  []string{},
		RemoteIPHeaders: []string{},
	})

	assert.Empty(t, config.TrustedProxies)
	assert.Empty(t, config.RemoteIPHeaders)
}

func TestConfigValidationKeepsDefaultTrustedProxies(t *testing.T) {
	config := isConfigValid(NewConfig())
	assert.Equal(t, []string{"0.0.0.0/0", "::/0"}, config.TrustedProxies)
}

func TestConfigWithTrustedProxiesOverridesDefaults(t *testing.T) {
	config := isConfigValid(NewConfig().WithTrustedProxies([]string{"10.0.0.0/8"}))
	assert.Equal(t, []string{"10.0.0.0/8"}, config.TrustedProxies)
}

func TestConfigWithCORSPolicyCloneInput(t *testing.T) {
	policy := com.CORSPolicy{
		Enabled:         true,
		AllowedOrigins:  []string{"https://a.example.com"},
		AllowedMethods:  []string{"GET"},
		AllowedHeaders:  []string{"Content-Type"},
		ExposeHeaders:   []string{"X-Request-Id"},
		MaxAgeSeconds:   60,
		AllowAllOrigins: false,
	}

	config := NewConfig().WithCORSPolicy(policy)
	policy.AllowedOrigins[0] = "https://changed.example.com"

	assert.NotNil(t, config.CORSPolicy)
	assert.Equal(t, []string{"https://a.example.com"}, config.CORSPolicy.AllowedOrigins)
}

func TestConfigWithCORSPolicyDisable(t *testing.T) {
	config := isConfigValid(NewConfig().WithCORSPolicy(com.CORSPolicy{Enabled: false}))
	assert.NotNil(t, config.CORSPolicy)
	assert.False(t, config.CORSPolicy.Enabled)
	assert.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, config.CORSPolicy.AllowedMethods)
}

func TestConfigValidationUsesDefaultCORSPolicyWhenNil(t *testing.T) {
	config := isConfigValid(&Config{
		Address:         "127.0.0.1",
		Port:            8080,
		TrustedProxies:  []string{},
		RemoteIPHeaders: []string{},
		CORSPolicy:      nil,
	})

	assert.NotNil(t, config.CORSPolicy)
	assert.True(t, config.CORSPolicy.Enabled)
	assert.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, config.CORSPolicy.AllowedMethods)
}

func TestIsOptionsValidNilUsesSafeDefaults(t *testing.T) {
	opts := isOptionsValid(nil)

	assert.NotNil(t, opts)
	assert.True(t, opts.healthCheck)
	assert.False(t, opts.pprof)
	assert.False(t, opts.swagger)
	assert.False(t, opts.metric)
	assert.False(t, opts.recReqBody)
	assert.False(t, opts.recRespBody)
}

func TestDebugOptionsEnableDebugFeatures(t *testing.T) {
	opts := DebugOptions()

	assert.True(t, opts.healthCheck)
	assert.True(t, opts.pprof)
	assert.True(t, opts.swagger)
	assert.True(t, opts.metric)
	assert.True(t, opts.recReqBody)
	assert.True(t, opts.recRespBody)
}

type respBodyProbeService struct {
	gotBody []byte
	gotErr  error
}

func (s *respBodyProbeService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/probe", func(c *gin.Context) {
		c.String(http.StatusOK, "hello orbit")
		s.gotBody, s.gotErr = uhttptool.GenerateResponseBody(c)
	})
}

type simpleWriteService struct{}

func (s *simpleWriteService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/probe", func(c *gin.Context) {
		c.String(http.StatusOK, "hello orbit")
	})
}

func newOptionsProbeEngine(tb testing.TB, options *Options, service Service, middlewares ...gin.HandlerFunc) *Engine {
	tb.Helper()

	config := NewConfig().
		WithRelease().
		WithPort(getFreePort(tb)).
		WithAccessLogEventFunc(benchmarkNoopLogEvent).
		WithRecoveryLogEventFunc(benchmarkNoopLogEvent).
		WithPrometheusRegistry(prometheus.NewRegistry())

	engine := NewEngine(config, options)
	if engine.initErr != nil {
		tb.Fatalf("engine init failed: %v", engine.initErr)
	}

	engine.RegisterService(service)
	for _, middleware := range middlewares {
		engine.RegisterMiddleware(middleware)
	}
	engine.Run()
	tb.Cleanup(engine.Stop)
	return engine
}

func TestResponseBodyRecordOptIn(t *testing.T) {
	t.Run("DisabledByDefault", func(t *testing.T) {
		probe := &respBodyProbeService{}
		engine := newOptionsProbeEngine(t, NewOptions(), probe)

		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		resp := httptest.NewRecorder()
		engine.ginSvr.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "hello orbit", resp.Body.String())
		assert.Equal(t, uhttptool.ErrorResponseBodyBufferNotFound, probe.gotErr)
		assert.Nil(t, probe.gotBody)
	})

	t.Run("EnabledByOption", func(t *testing.T) {
		probe := &respBodyProbeService{}
		engine := newOptionsProbeEngine(t, NewOptions().EnableRecordResponseBody(), probe)

		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		resp := httptest.NewRecorder()
		engine.ginSvr.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "hello orbit", resp.Body.String())
		assert.NoError(t, probe.gotErr)
		assert.Equal(t, []byte("hello orbit"), probe.gotBody)
	})

	t.Run("ReadInMiddlewareAfterNext", func(t *testing.T) {
		probe := &respBodyProbeService{}
		probeMiddleware := func(c *gin.Context) {
			c.Next()
			probe.gotBody, probe.gotErr = uhttptool.GenerateResponseBody(c)
		}
		engine := newOptionsProbeEngine(t, NewOptions().EnableRecordResponseBody(), &simpleWriteService{}, probeMiddleware)

		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		resp := httptest.NewRecorder()
		engine.ginSvr.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "hello orbit", resp.Body.String())
		assert.NoError(t, probe.gotErr)
		assert.Equal(t, []byte("hello orbit"), probe.gotBody)
	})
}

type reqBodyEchoService struct {
	mu             sync.Mutex
	capturedBuffer *bytes.Buffer
}

func (s *reqBodyEchoService) RegisterGroup(g *gin.RouterGroup) {
	g.POST("/echo", func(c *gin.Context) {
		body, err := uhttptool.GenerateRequestBody(c)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		if buf, exists := c.Get(com.RequestBodyBufferKey); exists {
			if b, ok := buf.(*bytes.Buffer); ok {
				s.mu.Lock()
				s.capturedBuffer = b
				s.mu.Unlock()
			}
		}
		c.Data(http.StatusOK, com.HttpHeaderJSONContentTypeValue, body)
	})
}

func TestAccessLoggerReturnsRequestBodyBufferToPool(t *testing.T) {
	service := &reqBodyEchoService{}
	engine := newOptionsProbeEngine(t, NewOptions().EnableRecordRequestBody(), service)

	payload := `{"seq":1,"data":"first"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(payload))
	req.Header.Set(com.HttpHeaderContentType, com.HttpHeaderJSONContentTypeValue)
	resp := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, payload, resp.Body.String())

	service.mu.Lock()
	captured := service.capturedBuffer
	service.mu.Unlock()
	require.NotNil(t, captured)
	assert.Zero(t, captured.Len())
}

func TestRecordRequestBodyAcrossConsecutiveRequests(t *testing.T) {
	var mu sync.Mutex
	recordedBodies := make([]string, 0, 2)

	config := NewConfig().
		WithRelease().
		WithPort(getFreePort(t)).
		WithAccessLogEventFunc(func(_ *logr.Logger, event *ulog.LogEvent) {
			mu.Lock()
			recordedBodies = append(recordedBodies, strings.Clone(event.ReqBody))
			mu.Unlock()
		}).
		WithRecoveryLogEventFunc(benchmarkNoopLogEvent).
		WithPrometheusRegistry(prometheus.NewRegistry())

	engine := NewEngine(config, NewOptions().EnableRecordRequestBody())
	require.NoError(t, engine.initErr)

	engine.RegisterService(&reqBodyEchoService{})
	engine.Run()
	t.Cleanup(engine.Stop)

	first := `{"seq":1,"data":"first"}`
	second := `{"seq":2,"data":"second"}`
	for _, payload := range []string{first, second} {
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(payload))
		req.Header.Set(com.HttpHeaderContentType, com.HttpHeaderJSONContentTypeValue)
		resp := httptest.NewRecorder()
		engine.ginSvr.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, payload, resp.Body.String())
	}

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, recordedBodies, 2)
	assert.Equal(t, first, recordedBodies[0])
	assert.Equal(t, second, recordedBodies[1])
}

func hasMetricFamily(registry *prometheus.Registry, name string) bool {
	families, err := registry.Gather()
	if err != nil {
		return false
	}
	for _, family := range families {
		if family.GetName() == name {
			return true
		}
	}
	return false
}

func TestMetricSharedRegistryAcrossEngines(t *testing.T) {
	registry := prometheus.NewRegistry()

	engine1 := NewEngine(NewConfig().WithRelease().WithPrometheusRegistry(registry), NewOptions().EnableMetric())
	require.NoError(t, engine1.initErr)
	require.True(t, engine1.IsMetricEnabled())

	var engine2 *Engine
	require.NotPanics(t, func() {
		engine2 = NewEngine(NewConfig().WithRelease().WithPrometheusRegistry(registry), NewOptions().EnableMetric())
	})
	require.NotNil(t, engine2)
	require.NoError(t, engine2.initErr)

	engine1.metric.IncRequestCount(http.MethodGet, "/probe", "200")
	assert.True(t, hasMetricFamily(registry, "orbit_http_requests_total"))

	engine1.Stop()
	assert.False(t, hasMetricFamily(registry, "orbit_http_requests_total"))

	assert.NotPanics(t, func() {
		engine2.Stop()
	})
	assert.False(t, engine1.IsRunning())
	assert.False(t, engine2.IsRunning())
}
