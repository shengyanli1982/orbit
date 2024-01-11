package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

func TestBodyBuffer(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Define a test middleware handler
	middlewareHandler := func(c *gin.Context) {
		buff := bytes.NewBuffer(make([]byte, 0, 2048))
		assert.Equal(t, buff, c.MustGet(com.RequestBodyBufferKey))
		assert.Equal(t, buff, c.MustGet(com.ResponseBodyBufferKey))
		c.Next()
	}

	// Use the BodyBuffer middleware
	router.Use(BodyBuffer(), middlewareHandler)

	// Define a test route
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response")
	})

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)

	// Create a test response recorder
	rec := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(rec, req)
}
