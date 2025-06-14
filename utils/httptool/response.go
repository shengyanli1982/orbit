package httptool

import (
	"bytes"
	"errors"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// 表示响应体缓冲区为空的错误
var ErrorResponseBodyBufferEmpty = errors.New("response body buffer is empty")

// 表示未找到请求体缓冲区的错误
var ErrorRequestBodyBufferNotFound = errors.New("request body buffer not found")

// 从 gin.Context 中生成响应体
func GenerateResponseBody(context *gin.Context) ([]byte, error) {
	// 从上下文中获取响应体缓冲区
	if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
		respBodyBuffer := buffer.(*bytes.Buffer)

		// 检查缓冲区是否为空
		if respBodyBuffer.Len() <= 0 {
			return nil, ErrorResponseBodyBufferEmpty
		}

		// 返回缓冲区中的字节数据
		return respBodyBuffer.Bytes(), nil
	} else {
		// 如果未找到缓冲区，返回错误
		return nil, ErrorRequestBodyBufferNotFound
	}
}
