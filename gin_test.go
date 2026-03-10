package orbit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
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
	assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
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

	engine.Run()

	assert.False(t, engine.IsRunning())
	assert.Nil(t, engine.httpSvr)
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
