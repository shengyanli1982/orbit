package httptool

import (
	"bytes"
	"errors"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// ErrorResponseBodyBufferEmpty 表示响应体缓冲区为空的错误。
// ErrorResponseBodyBufferEmpty indicates that the response body buffer is empty.
var ErrorResponseBodyBufferEmpty = errors.New("response body buffer is empty")

// ErrorRequestBodyBufferNotFound 表示未找到请求体缓冲区的错误。
// ErrorRequestBodyBufferNotFound indicates that the request body buffer was not found.
var ErrorRequestBodyBufferNotFound = errors.New("request body buffer not found")

// GenerateResponseBody 函数从 gin.Context 中生成响应体。
// The GenerateResponseBody function generates a response body from the gin.Context.
func GenerateResponseBody(context *gin.Context) ([]byte, error) {
	// 从上下文中获取响应体缓冲区
	// Get the response body buffer from the context
	if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
		respBodyBuffer := buffer.(*bytes.Buffer)

		// 检查缓冲区是否为空
		// Check if the buffer is empty
		if respBodyBuffer.Len() <= 0 {
			return nil, ErrorResponseBodyBufferEmpty
		}

		// 返回缓冲区中的字节数据
		// Return the bytes from the buffer
		return respBodyBuffer.Bytes(), nil
	} else {
		// 如果未找到缓冲区，返回错误
		// Return error if buffer is not found
		return nil, ErrorRequestBodyBufferNotFound
	}
}
