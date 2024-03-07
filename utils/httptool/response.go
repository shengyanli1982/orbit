package httptool

import (
	"bytes"
	"errors"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

var (
	// ErrorResponseBodyBufferEmpty 是当响应体缓冲区为空时返回的错误。
	// ErrorResponseBodyBufferEmpty is the error returned when the response body buffer is empty.
	ErrorResponseBodyBufferEmpty = errors.New("response body buffer is empty")
	// ErrorRequestBodyBufferNotFound 是当请求体缓冲区未找到时返回的错误。
	// ErrorRequestBodyBufferNotFound is the error returned when the request body buffer is not found.
	ErrorRequestBodyBufferNotFound = errors.New("request body buffer not found")
)

// GenerateResponseBody 函数从响应体缓冲区生成响应体。
// The GenerateResponseBody function generates the response body from the response body buffer.
func GenerateResponseBody(context *gin.Context) ([]byte, error) {
	// 从上下文中获取响应体缓冲区
	// Get the response body buffer from the context
	if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
		// 从缓冲区获取响应体内容
		// Get the response body content from the buffer
		respBodyBuffer := buffer.(*bytes.Buffer)

		// 检查响应体缓冲区是否为空
		// Check if the response body buffer is empty
		if respBodyBuffer.Len() <= 0 {
			// 如果响应体缓冲区为空，返回错误
			// If the response body buffer is empty, return an error
			return nil, ErrorResponseBodyBufferEmpty
		}

		// 如果响应体缓冲区不为空，返回响应体内容
		// If the response body buffer is not empty, return the response body content
		return respBodyBuffer.Bytes(), nil
	} else {
		// 如果响应体缓冲区未找到，返回错误
		// If the response body buffer is not found, return an error
		return nil, ErrorRequestBodyBufferNotFound
	}
}
