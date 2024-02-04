package httptool

import (
	"bytes"
	"errors"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

var (
	// ErrorResponseBodyBufferEmpty 是响应体缓冲区为空时的错误。
	// ErrorResponseBodyBufferEmpty is the error when the response body buffer is empty.
	ErrorResponseBodyBufferEmpty = errors.New("response body buffer is empty")

	// ErrorRequestBodyBufferNotFound 是请求体缓冲区未找到时的错误。
	// ErrorRequestBodyBufferNotFound is the error when the request body buffer is not found.
	ErrorRequestBodyBufferNotFound = errors.New("request body buffer not found")
)

// GenerateResponseBody 从响应体缓冲区生成响应体。
// GenerateResponseBody generates the response body from the response body buffer.
func GenerateResponseBody(context *gin.Context) ([]byte, error) {
	// 获得上下文中的响应体缓冲区
	// Get the response body buffer from the context
	if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
		// 从缓冲区中获取响应体内容
		// Get the response body content from the buffer
		respBodyBuffer := buffer.(*bytes.Buffer)

		// 检查响应体缓冲区是否为空
		// Check if the response body buffer is empty
		if respBodyBuffer.Len() <= 0 {
			return nil, ErrorResponseBodyBufferEmpty
		}

		// 返回响应体内容
		// Return the response body content
		return respBodyBuffer.Bytes(), nil
	} else {
		// 如果未找到响应体缓冲区，返回错误
		// If the response body buffer is not found, return an error
		return nil, ErrorRequestBodyBufferNotFound
	}
}
