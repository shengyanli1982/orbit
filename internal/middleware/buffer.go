package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// BodyBuffer 返回一个 Gin 中间件函数，用于处理请求和响应的缓冲。
// BodyBuffer returns a Gin middleware function for handling request and response buffering.
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 从缓冲池中获取请求体缓冲区和响应体缓冲区
		// Get request and response body buffers from the buffer pool
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		respBodyBuffer := com.ResponseBodyBufferPool.Get()

		// 将缓冲区存储在上下文中，以便后续使用
		// Store buffers in the context for later use
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// 创建一个新的响应体写入器，包装原始的响应写入器
		// Create a new response body writer that wraps the original writer
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)
		originalWriter := context.Writer
		context.Writer = bufferedWriter

		// 执行后续的中间件和处理函数
		// Execute subsequent middleware and handlers
		context.Next()

		// 恢复原始的响应写入器并重置缓冲的写入器
		// Restore the original writer and reset the buffered writer
		context.Writer = originalWriter
		bufferedWriter.Reset()

		// 清理请求体缓冲区并将其返回到池中
		// Clean up request body buffer and return it to the pool
		if reqBuffer, ok := context.Get(com.RequestBodyBufferKey); ok {
			if buf, ok := reqBuffer.(*bytes.Buffer); ok {
				buf.Reset()
				com.RequestBodyBufferPool.Put(buf)
			}
			// 从上下文中移除请求体缓冲区引用
			// Remove the request body buffer reference from the context
			context.Set(com.RequestBodyBufferKey, nil)
		}

		// 清理响应体缓冲区并将其返回到池中
		// Clean up response body buffer and return it to the pool
		if respBuffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
			if buf, ok := respBuffer.(*bytes.Buffer); ok {
				buf.Reset()
				com.ResponseBodyBufferPool.Put(buf)
			}
			// 从上下文中移除响应体缓冲区引用
			// Remove the response body buffer reference from the context
			context.Set(com.ResponseBodyBufferKey, nil)
		}
	}
}
