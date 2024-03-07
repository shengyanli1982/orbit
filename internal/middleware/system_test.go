package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	ilog "github.com/shengyanli1982/orbit/internal/log"
	"github.com/shengyanli1982/orbit/utils/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCors(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Add the Cors middleware to the router
	router.Use(Cors())

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert that the response body contains the expected message
	header := recorder.Header()
	assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "*")
	assert.Equal(t, header.Get("Access-Control-Allow-Methods"), "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	assert.Equal(t, header.Get("Access-Control-Allow-Headers"), "*")
	assert.Equal(t, header.Get("Access-Control-Expose-Headers"), "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
	assert.Equal(t, header.Get("Access-Control-Allow-Credentials"), "true")
	assert.Equal(t, header.Get("Access-Control-Max-Age"), "172800")
}

func TestRecovery(t *testing.T) {
	// Create a test Gin router
	logger := zap.NewExample().Sugar()

	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		panic("test panic")
	}

	// Add the Recovery middleware to the router
	router.Use(Recovery(logger, ilog.DefaultRecoveryEventFunc))

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 500
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "[500] http server internal error, method: GET, path: /test", recorder.Body.String())
}

func TestAccessLogger(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Create a test logger
	logger := zap.NewExample().Sugar()

	// Create a test log event function
	logEventFunc := func(logger *zap.SugaredLogger, event *log.LogEvent) {
		logger.Info(event)
	}

	// Add the AccessLogger middleware to the router
	router.Use(AccessLogger(logger, logEventFunc, true))

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "{\"message\":\"OK\"}", recorder.Body.String())
}
