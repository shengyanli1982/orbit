package httptool

import (
	"bytes"
	"errors"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

var (
	// ErrorResponseBodyBufferEmpty is the error when the response body buffer is empty.
	ErrorResponseBodyBufferEmpty = errors.New("response body buffer is empty")
	// ErrorRequestBodyBufferNotFound is the error when the request body buffer is not found.
	ErrorRequestBodyBufferNotFound = errors.New("request body buffer not found")
)

// GenerateResponseBody generates the response body from the response body buffer.
func GenerateResponseBody(context *gin.Context) ([]byte, error) {
	// Get the response body buffer from the context
	if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
		// Get the response body content from the buffer
		respBodyBuffer := buffer.(*bytes.Buffer)
		// Check if the response body buffer is empty
		if respBodyBuffer.Len() <= 0 {
			return nil, ErrorResponseBodyBufferEmpty
		}
		return respBodyBuffer.Bytes(), nil
	} else {
		// If the response body buffer is not found, return an error
		return nil, ErrorRequestBodyBufferNotFound
	}
}
