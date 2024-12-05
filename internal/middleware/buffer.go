package middleware

import (
	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// BodyBuffer 返回一个 Gin 中间件函数，用于处理请求和响应的缓冲。
// BodyBuffer returns a Gin middleware function for handling request and response buffering.
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 从缓冲池中获取缓冲区
		// Get buffers from the pool
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		respBodyBuffer := com.ResponseBodyBufferPool.Get()

		// 预先清空缓冲区，确保没有残留数据
		// Clear buffers to ensure no residual data
		reqBodyBuffer.Reset()
		respBodyBuffer.Reset()

		defer func() {
			// 使用 defer 确保缓冲区一定会被清理和归还到池中
			// Use defer to ensure buffers are cleaned up and returned to the pool
			reqBodyBuffer.Reset()
			respBodyBuffer.Reset()
			com.RequestBodyBufferPool.Put(reqBodyBuffer)
			com.ResponseBodyBufferPool.Put(respBodyBuffer)
		}()

		// 将缓冲区存储在上下文中
		// Store buffers in the context
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// 创建并设置响应写入器
		// Create and set response writer
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)
		originalWriter := context.Writer
		context.Writer = bufferedWriter

		// 执行后续的中间件和处理函数
		// Execute subsequent middleware and handlers
		context.Next()

		// 恢复原始的响应写入器
		// Restore the original writer
		context.Writer = originalWriter
		bufferedWriter.Reset()

		// 清除上下文中的引用
		// Clear references from context
		context.Set(com.RequestBodyBufferKey, nil)
		context.Set(com.ResponseBodyBufferKey, nil)
	}
}
