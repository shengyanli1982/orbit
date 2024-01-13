package httptool

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

func TestGenerateResponseBody(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	t.Run("NonEmptyResponseBodyBuffer", func(t *testing.T) {
		// Create a response body buffer with some content
		respBodyBuffer := bytes.NewBuffer([]byte("Hello, World!"))

		// Set the response body buffer in the context
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// Call the GenerateResponseBody function
		body, err := GenerateResponseBody(context)

		// Check if the error is nil
		assert.Nil(t, err)

		// Check if the body is as expected
		expectedBody := []byte("Hello, World!")
		assert.Equal(t, expectedBody, body)
	})
}

func TestGenerateResponseBodyEmpty(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	t.Run("EmptyResponseBodyBuffer", func(t *testing.T) {
		// Set an empty response body buffer in the context
		context.Set(com.ResponseBodyBufferKey, bytes.NewBuffer([]byte{}))

		// Call the GenerateResponseBody function
		body, err := GenerateResponseBody(context)

		// Check if the error is as expected
		assert.Equal(t, ErrorResponseBodyBufferEmpty, err)

		// Check if the body is nil
		assert.Nil(t, body)
	})
}

func TestGenerateResponseBodyBufferNotFound(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	t.Run("ResponseBodyBufferNotFound", func(t *testing.T) {
		// Call the GenerateResponseBody function without setting the response body buffer in the context
		body, err := GenerateResponseBody(context)

		// Check if the error is as expected
		assert.Equal(t, ErrorRequestBodyBufferNotFound, err)

		// Check if the body is nil
		assert.Nil(t, body)
	})
}
