package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit/utils/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
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

func TestAccessLogger(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewZapLogger(zapcore.AddSync(buff)).GetLogrLogger()

	// Create a test log event function
	logEventFunc := func(logger *logr.Logger, event *log.LogEvent) {
		logger.Info(event.Message)
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

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server access log", "buffer should contain the message")
}

func TestLogrAccessLogger(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewLogrLogger(buff).GetLogrLogger()

	// Create a test log event function
	logEventFunc := func(logger *logr.Logger, event *log.LogEvent) {
		logger.Info(event.Message, "code", event.Code, "method", event.Method, "path", event.Path)
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

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server access log", "buffer should contain the message")
}

func TestRecovery(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		panic("test panic")
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewZapLogger(zapcore.AddSync(buff)).GetLogrLogger()

	// Add the Recovery middleware to the router
	router.Use(Recovery(logger, log.DefaultRecoveryEventFunc))

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

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server recovery from panic", "buffer should contain the message")
}

func TestLogrRecovery(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		panic("test panic")
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewLogrLogger(buff).GetLogrLogger()

	// Add the Recovery middleware to the router
	router.Use(Recovery(logger, log.DefaultRecoveryEventFunc))

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

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server recovery from panic", "buffer should contain the message")
}
