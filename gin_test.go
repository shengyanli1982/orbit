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

func TestNewEngine(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := &Options{
		forwordByClientIp: true,
		trailingSlash:     true,
		fixedPath:         true,
		swagger:           true,
		pprof:             true,
		metric:            true,
	}

	// Call the NewEngine function
	engine := NewEngine(config, options)

	// Run the engine
	engine.Run()
	defer engine.Stop()

	// Assert that the engine is not nil
	assert.NotNil(t, engine)

	// Assert that the engine's running field is false
	assert.True(t, engine.running)

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
	options := &Options{
		forwordByClientIp: true,
		trailingSlash:     true,
		fixedPath:         true,
		swagger:           true,
		pprof:             true,
		metric:            true,
	}

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
	options := &Options{
		forwordByClientIp: true,
		trailingSlash:     true,
		fixedPath:         true,
		swagger:           true,
		pprof:             true,
		metric:            true,
	}

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
	options := &Options{
		forwordByClientIp: true,
		trailingSlash:     true,
		fixedPath:         true,
		swagger:           true,
		pprof:             true,
		metric:            true,
	}

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

func TestRegisterMiddleware(t *testing.T) {
	// Create a new Config
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create a new Options
	options := &Options{
		forwordByClientIp: true,
		trailingSlash:     true,
		fixedPath:         true,
		swagger:           true,
		pprof:             true,
		metric:            true,
	}

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
