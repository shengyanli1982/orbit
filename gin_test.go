package orbit

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type emptyBodyService struct{}

func (s *emptyBodyService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/empty", func(ctx *gin.Context) {})
}

type clientIPService struct{}

func (s *clientIPService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/client-ip", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, ctx.ClientIP())
	})
}

func TestNewEngine(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := NewOptions()

	// Call the NewEngine function
	engine := NewEngine(config, options)

	// Run the engine
	engine.Run()
	defer engine.Stop()

	// Assert that the engine is not nil
	assert.NotNil(t, engine)

	// Assert that the engine's running field is true (使用 IsRunning() 方法)
	assert.True(t, engine.IsRunning())

	// Assert that the engine's endpoint matches the expected value
	assert.Equal(t, "localhost:8080", engine.endpoint)

	// Assert that the engine's config matches the expected value
	assert.Equal(t, config, engine.config)

	// Assert that the engine's opts matches the expected value
	assert.Equal(t, options, engine.opts)

	// Assert that the engine's handlers slice is empty
	assert.Empty(t, engine.handlers)

	// Assert that the engine's services slice is empty
	assert.Empty(t, engine.services)

	// Assert that the engine's ctx is not nil
	assert.NotNil(t, engine.ctx)

	// Assert that the engine's cancel is not nil
	assert.NotNil(t, engine.cancel)

	// Assert that the engine's ginSvr is not nil
	assert.NotNil(t, engine.ginSvr)

	// Assert that the engine's root is not nil
	assert.NotNil(t, engine.root)
}

func TestNewEngineNoRoute(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := NewOptions()

	// Call the NewEngine function
	engine := NewEngine(config, options)

	// Run the engine
	engine.Run()
	defer engine.Stop()

	// Create a new HTTP request
	req, _ := http.NewRequest(http.MethodGet, "/not-found", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	engine.ginSvr.ServeHTTP(recorder, req)

	// Assert that the response status code is 404
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	// Assert that the response body matches the expected value
	assert.Equal(t, "[404] http request route mismatch, method: GET, path: /not-found", recorder.Body.String())
}

func TestNewEngineNoMethod(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := NewOptions()

	// Call the NewEngine function
	engine := NewEngine(config, options)

	// Run the engine
	engine.Run()
	defer engine.Stop()

	// Create a new HTTP request
	req, _ := http.NewRequest(http.MethodPost, "/ping", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	engine.ginSvr.ServeHTTP(recorder, req)

	// Assert that the response status code is 405
	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)

	// Assert that the response body matches the expected value
	assert.Equal(t, "[405] http request method not allowed, method: POST, path: /ping", recorder.Body.String())
}

func TestEngineNoRouteAndNoMethodBodies(t *testing.T) {
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}
	engine := NewEngine(config, NewOptions())
	engine.root.POST("/only-post", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	engine.Run()
	defer engine.Stop()

	cases := []struct {
		method     string
		path       string
		expectCode int
		expectBody string
	}{
		{
			method:     http.MethodDelete,
			path:       "/missing/deep/path",
			expectCode: http.StatusNotFound,
			expectBody: "[404] http request route mismatch, method: DELETE, path: /missing/deep/path",
		},
		{
			method:     http.MethodGet,
			path:       "/only-post",
			expectCode: http.StatusMethodNotAllowed,
			expectBody: "[405] http request method not allowed, method: GET, path: /only-post",
		},
		{
			method:     http.MethodPut,
			path:       com.HealthCheckURLPath,
			expectCode: http.StatusMethodNotAllowed,
			expectBody: "[405] http request method not allowed, method: PUT, path: /ping",
		},
	}

	for _, tc := range cases {
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		recorder := httptest.NewRecorder()
		engine.ginSvr.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expectCode, recorder.Code, "method=%s path=%s", tc.method, tc.path)
		assert.Equal(t, tc.expectBody, recorder.Body.String(), "method=%s path=%s", tc.method, tc.path)
	}
}

func TestNewEngineHealthCheck(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := NewOptions()

	// Call the NewEngine function
	engine := NewEngine(config, options)

	// Run the engine
	engine.Run()
	defer engine.Stop()

	// Create a new HTTP request
	req, _ := http.NewRequest(http.MethodGet, com.HealthCheckURLPath, nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	engine.ginSvr.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert that the response body matches the expected value
	assert.Equal(t, com.RequestOK, recorder.Body.String())
}

func TestEngineDefaultCorsPolicyIsConservative(t *testing.T) {
	engine := NewEngine(NewConfig(), NewOptions())
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, com.HealthCheckURLPath, nil)
	req.Header.Set("Origin", "https://app.example.com")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
}

func TestEngineCorsPolicyCanAllowAllOrigins(t *testing.T) {
	config := NewConfig().WithCORSPolicy(com.CORSPolicy{
		Enabled:          true,
		AllowAllOrigins:  true,
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAgeSeconds:    120,
	})
	engine := NewEngine(config, NewOptions())
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, com.HealthCheckURLPath, nil)
	req.Header.Set("Origin", "https://app.example.com")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
}

func TestRegisterMiddleware(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := NewOptions()

	// Call the NewEngine function
	engine := NewEngine(config, options)

	// Create a new middleware handler
	middlewareHandler := func(c *gin.Context) {
		c.Next()
		c.String(http.StatusOK, "middleware")
	}

	// Register the middleware handler
	engine.RegisterMiddleware(middlewareHandler)

	// Run the engine
	engine.RegisterService(&emptyBodyService{})

	// Run the engine
	engine.Run()
	defer engine.Stop()

	// Create a new HTTP request
	req, _ := http.NewRequest(http.MethodGet, "/empty", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	engine.ginSvr.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert that the response body matches the expected value
	assert.Equal(t, "middleware", recorder.Body.String())
}

func TestPprofServiceRoute(t *testing.T) {
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}
	options := NewOptions().EnablePProf()
	engine := NewEngine(config, options)

	req, _ := http.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	req, _ = http.NewRequest(http.MethodGet, "/debug/pprof/debug/pprof/", nil)
	recorder = httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestClientIPForwardedDisabledIgnoresHeader(t *testing.T) {
	engine := NewEngine(NewConfig(), NewOptions())
	engine.RegisterService(&clientIPService{})
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "10.1.2.3", recorder.Body.String())
}

func TestClientIPForwardedEnabledUsesHeaderByDefaultTrustedProxies(t *testing.T) {
	engine := NewEngine(NewConfig(), NewOptions().EnableForwardedByClientIp())
	engine.RegisterService(&clientIPService{})
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "1.2.3.4", recorder.Body.String())
}

func TestClientIPForwardedEnabledTrustedProxyUsesHeader(t *testing.T) {
	config := NewConfig().WithTrustedProxies([]string{"10.0.0.0/8"})
	options := NewOptions().EnableForwardedByClientIp()
	engine := NewEngine(config, options)
	engine.RegisterService(&clientIPService{})
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "1.2.3.4", recorder.Body.String())
}

func TestClientIPForwardedEnabledUntrustedProxyIgnoresHeader(t *testing.T) {
	config := NewConfig().WithTrustedProxies([]string{"192.168.0.0/16"})
	options := NewOptions().EnableForwardedByClientIp()
	engine := NewEngine(config, options)
	engine.RegisterService(&clientIPService{})
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "10.1.2.3", recorder.Body.String())
}

func TestClientIPForwardedEnabledWithNoTrustedProxiesIgnoresHeader(t *testing.T) {
	config := NewConfig().WithTrustedProxies([]string{})
	options := NewOptions().EnableForwardedByClientIp()
	engine := NewEngine(config, options)
	engine.RegisterService(&clientIPService{})
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "10.1.2.3", recorder.Body.String())
}

func TestClientIPForwardedEnabledUsesConfiguredHeaderOrder(t *testing.T) {
	config := NewConfig().
		WithTrustedProxies([]string{"10.0.0.0/8"}).
		WithRemoteIPHeaders([]string{"X-Real-IP"})
	options := NewOptions().EnableForwardedByClientIp()
	engine := NewEngine(config, options)
	engine.RegisterService(&clientIPService{})
	engine.Run()
	defer engine.Stop()

	req, _ := http.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "5.6.7.8")
	recorder := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "5.6.7.8", recorder.Body.String())
}

func TestRunFailFastWhenForwardedEnabledAndTrustedProxiesInvalid(t *testing.T) {
	config := NewConfig().WithTrustedProxies([]string{"invalid-cidr"})
	options := NewOptions().EnableForwardedByClientIp()
	engine := NewEngine(config, options)

	assert.Error(t, engine.initErr)
	assert.Error(t, engine.GetInitError())
	assert.Equal(t, engine.initErr, engine.GetInitError())

	engine.Run()

	assert.False(t, engine.IsRunning())
	assert.Nil(t, engine.httpSvr)
}

func TestGetInitErrorNilWhenInitSucceeds(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())

	assert.NoError(t, engine.GetInitError())
}

func TestRunNotFailWhenForwardedDisabledEvenIfTrustedProxiesInvalid(t *testing.T) {
	config := NewConfig().WithTrustedProxies([]string{"invalid-cidr"})
	options := NewOptions()
	engine := NewEngine(config, options)

	assert.NoError(t, engine.initErr)

	engine.Run()
	defer engine.Stop()

	assert.True(t, engine.IsRunning())
}

func getFreePort(tb testing.TB) uint16 {
	tb.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		tb.Fatalf("failed to allocate a free port: %v", err)
	}
	port := uint16(listener.Addr().(*net.TCPAddr).Port)
	if err := listener.Close(); err != nil {
		tb.Fatalf("failed to close the probe listener: %v", err)
	}
	return port
}

func TestEngineConcurrentRunSingleWinner(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())
	engine.RegisterService(&emptyBodyService{})

	const workers = 10
	start := make(chan struct{})
	var runWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		runWG.Add(1)
		go func() {
			defer runWG.Done()
			<-start
			engine.Run()
		}()
	}
	close(start)
	runWG.Wait()

	time.Sleep(200 * time.Millisecond)

	assert.True(t, engine.IsRunning())
	assert.Nil(t, engine.GetRunError())

	engine.Stop()
	assert.False(t, engine.IsRunning())
}

func TestEngineRunAfterStopRejected(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())

	engine.Run()
	assert.True(t, engine.IsRunning())
	engine.Stop()
	assert.False(t, engine.IsRunning())

	done := make(chan struct{})
	go func() {
		engine.Run()
		engine.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run/Stop after stop timed out, possible deadlock")
	}

	assert.False(t, engine.IsRunning())
	assert.Nil(t, engine.GetRunError())
}

func TestEngineStopBeforeRun(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())

	engine.Stop()
	assert.False(t, engine.IsRunning())

	engine.Run()
	assert.False(t, engine.IsRunning())
	assert.Nil(t, engine.GetRunError())
}

func TestEngineDoubleStopIsIdempotent(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())

	engine.Run()
	assert.True(t, engine.IsRunning())

	engine.Stop()
	assert.False(t, engine.IsRunning())
	engine.Stop()
	assert.False(t, engine.IsRunning())
}

func TestEngineListenFailureReachesTerminalState(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to occupy a port: %v", err)
	}
	defer listener.Close()
	port := uint16(listener.Addr().(*net.TCPAddr).Port)

	engine := NewEngine(NewConfig().WithRelease().WithPort(port), NewOptions())
	engine.Run()

	deadline := time.Now().Add(5 * time.Second)
	for engine.GetRunError() == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	assert.Error(t, engine.GetRunError())
	assert.False(t, engine.IsRunning())

	engine.Run()
	assert.False(t, engine.IsRunning())
}

func TestEngineListenFailureUnregistersMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to occupy a port: %v", err)
	}
	defer listener.Close()
	port := uint16(listener.Addr().(*net.TCPAddr).Port)

	engine := NewEngine(NewConfig().WithRelease().WithPort(port).WithPrometheusRegistry(registry), NewOptions().EnableMetric())
	require.NoError(t, engine.initErr)
	require.True(t, engine.IsMetricEnabled())

	engine.metric.IncRequestCount(http.MethodGet, "/probe", "200")
	require.True(t, hasMetricFamily(registry, "orbit_http_requests_total"))

	engine.Run()

	deadline := time.Now().Add(5 * time.Second)
	for engine.GetRunError() == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	assert.Error(t, engine.GetRunError())
	assert.False(t, engine.IsRunning())
	assert.False(t, hasMetricFamily(registry, "orbit_http_requests_total"))
}

func TestEngineRegisterAfterRunRejected(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())

	engine.Run()
	defer engine.Stop()
	assert.True(t, engine.IsRunning())

	engine.RegisterService(&emptyBodyService{})
	engine.RegisterMiddleware(func(c *gin.Context) { c.Next() })

	assert.Empty(t, engine.services)
	assert.Empty(t, engine.handlers)
}

// TestStopWithContextCustomDeadline 验证 StopWithContext 尊重调用者提供的短超时上下文，
// 而非使用配置的 ShutdownTimeout（默认 10s）：在有慢请求在途时，ctx 超时后应快速返回。
func TestStopWithContextCustomDeadline(t *testing.T) {
	port := getFreePort(t)
	engine := NewEngine(NewConfig().WithRelease().WithPort(port), NewOptions())
	require.NoError(t, engine.GetInitError())

	inHandler := make(chan struct{})
	var handlerDone atomic.Bool
	engine.RegisterService(NewHttpService(func(g *gin.RouterGroup) {
		g.GET("/slow", func(c *gin.Context) {
			close(inHandler)
			// 模拟慢请求：远长于 ctx 超时
			time.Sleep(5 * time.Second)
			handlerDone.Store(true)
			c.String(http.StatusOK, "done")
		})
	}))
	engine.Run()
	defer engine.Stop()

	// 等待服务器就绪后发起慢请求
	client := &http.Client{Timeout: 10 * time.Second}
	reqDone := make(chan struct{})
	go func() {
		defer close(reqDone)
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/slow", port))
		if err != nil {
			return // ctx 超时后连接会被强制关闭，预期错误
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	<-inHandler // 确保请求已进入 handler

	// 使用 500ms 短超时调用 StopWithContext
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	engine.StopWithContext(ctx)
	elapsed := time.Since(start)

	// StopWithContext 应在 ctx 超时后快速返回（远低于默认 10s）
	assert.Less(t, elapsed, 3*time.Second,
		"StopWithContext should return shortly after ctx deadline, took %v", elapsed)
	assert.False(t, engine.IsRunning())

	<-reqDone
	_ = handlerDone.Load() // 避免 unused variable
}

// TestStopWithContextCancelledContext 验证传入已取消的 ctx 时，
// StopWithContext 不会阻塞，应快速返回。
func TestStopWithContextCancelledContext(t *testing.T) {
	port := getFreePort(t)
	engine := NewEngine(NewConfig().WithRelease().WithPort(port), NewOptions())
	require.NoError(t, engine.GetInitError())

	engine.Run()
	defer engine.Stop()

	// 短暂等待确保 Run 的装配段完成
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	start := time.Now()
	engine.StopWithContext(ctx)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 2*time.Second,
		"StopWithContext with cancelled ctx should return quickly, took %v", elapsed)
	assert.False(t, engine.IsRunning())
}

// TestStopWithContextBackground 验证传入 context.Background()（无 deadline）时，
// 在无活跃连接的正常情况下，StopWithContext 应快速完成关闭。
func TestStopWithContextBackground(t *testing.T) {
	port := getFreePort(t)
	engine := NewEngine(NewConfig().WithRelease().WithPort(port), NewOptions())
	require.NoError(t, engine.GetInitError())

	engine.RegisterService(&emptyBodyService{})
	engine.Run()
	defer engine.Stop()

	// 短暂等待确保 Run 的装配段完成
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	engine.StopWithContext(context.Background())
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 3*time.Second,
		"StopWithContext with background ctx and no active connections should return quickly, took %v", elapsed)
	assert.False(t, engine.IsRunning())
}
