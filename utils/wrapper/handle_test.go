package wrapper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWrapHandlerFuncToGin(t *testing.T) {
	// Create a test Gin router
	router := gin.Default()

	// Create a test HTTP request
	req, _ := http.NewRequest("GET", "/test", nil)

	// Create a test HTTP response recorder
	recorder := httptest.NewRecorder()

	// Create a test handler function
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	// Wrap the handler function to Gin handler function
	ginHandlerFunc := WrapHandlerFuncToGin(handlerFunc)

	// Set the Gin handler function to the router
	router.GET("/test", ginHandlerFunc)

	// Perform the test request
	router.ServeHTTP(recorder, req)

	// Assert the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert the response body
	assert.Equal(t, "OK", recorder.Body.String())
}

func TestWrapHandlerToGin(t *testing.T) {
	// Create a test Gin router
	router := gin.Default()

	// Create a test HTTP request
	req, _ := http.NewRequest("GET", "/test", nil)

	// Create a test HTTP response recorder
	recorder := httptest.NewRecorder()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap the handler to Gin handler function
	ginHandlerFunc := WrapHandlerToGin(handler)

	// Set the Gin handler function to the router
	router.GET("/test", ginHandlerFunc)

	// Perform the test request
	router.ServeHTTP(recorder, req)

	// Assert the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert the response body
	assert.Equal(t, "OK", recorder.Body.String())
}
