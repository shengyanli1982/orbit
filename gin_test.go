package orbit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

type mockService struct{}

func (s *mockService) SomeMethod(c *gin.Context) {
	c.String(http.StatusOK, "mock")
}

func (s *mockService) RegisterGroup(routerGroup *gin.RouterGroup) {
	group := routerGroup.Group("/mock")
	group.GET(com.EmptyURLPath, s.SomeMethod)
}

func TestNewEngine(t *testing.T) {
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	options := &Options{
		forwordByClientIp: true,
		trailingSlash:     true,
		fixedPath:         true,
		swagger:           true,
		pprof:             true,
		metric:            true,
	}

	engine := NewEngine(config, options)

	assert.NotNil(t, engine)
	assert.False(t, engine.running)
	assert.Equal(t, "localhost:8080", engine.endpoint)
	assert.Equal(t, config, engine.config)
	assert.Equal(t, options, engine.opts)
	assert.NotNil(t, engine.ctx)
	assert.NotNil(t, engine.cancel)
	assert.NotNil(t, engine.ginSvr)
	assert.NotNil(t, engine.root)
	assert.NotNil(t, engine.metric)
}

func TestNewEngineNoRoute(t *testing.T) {
	// Create configuration
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create options
	options := &Options{}

	// Create engine
	engine := NewEngine(config, options)

	// Start the engine
	engine.Run()
	defer engine.Stop()

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notfound", nil)
	engine.ginSvr.ServeHTTP(w, req)

	// Assert that the response status code is 404
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "[404] http request route mismatch, method: GET, path: /notfound", w.Body.String())
}

// func TestNewEngineNoMethod(t *testing.T) {
// 	// Create configuration
// 	config := &Config{
// 		Address:     "localhost",
// 		Port:        8080,
// 		ReleaseMode: true,
// 	}

// 	// Create options
// 	options := &Options{}

// 	// Create engine
// 	engine := NewEngine(config, options)

// 	// Define a test route
// 	engine.root.GET("/notallowed", func(ctx *gin.Context) {})

// 	// Start the engine
// 	engine.Run()
// 	defer engine.Stop()

// 	// Create a test request
// 	w := httptest.NewRecorder()

// 	// Test OPTIONS method
// 	req, _ := http.NewRequest("GET", "/notallowed", nil)
// 	req.Header.Set("Access-Control-Request-Method", "XXXX")

// 	// Perform the request
// 	engine.ginSvr.ServeHTTP(w, req)

// 	// Assert that the response status code is 405
// 	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
// 	assert.Equal(t, "[405] http request method not allowed, method: POST, path: /notallowed", w.Body.String())
// }

func TestEngine_RegisterMiddleware(t *testing.T) {
	// Create configuration
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create options
	options := &Options{}

	// Create engine
	engine := NewEngine(config, options)

	// Define a test middleware handler
	middlewareHandler := func(c *gin.Context) {
		c.Set("middleware", true)
		c.Next()
		assert.True(t, c.GetBool("middleware"))
	}

	// Register the middleware
	engine.RegisterMiddleware(middlewareHandler)

	// Start the engine
	engine.Run()
	defer engine.Stop()
}

func TestEngine_RegisterService(t *testing.T) {
	// Create configuration
	config := &Config{
		Address:     "localhost",
		Port:        8080,
		ReleaseMode: true,
	}

	// Create options
	options := &Options{}

	// Create engine
	engine := NewEngine(config, options)

	// Define a test service
	service := &mockService{}

	// Register the service
	engine.RegisterService(service)

	// Start the engine
	engine.Run()
	defer engine.Stop()

	// Assert that the service is registered
	assert.Contains(t, engine.services, service)

	// Create a test request
	req, _ := http.NewRequest("GET", "/mock", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	engine.ginSvr.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "mock", recorder.Body.String())
}
