package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// BodyBuffer 是一个中间件，用于缓冲请求和响应体。
// BodyBuffer is a middleware that buffers the request and response bodies.
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 设置请求和响应体缓冲区
		// Set request and response body buffers
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		respBodyBuffer := com.ResponseBodyBufferPool.Get()
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// 创建一个新的缓冲区响应写入器
		// Create a new buffer response writer
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)

		// 替换响应写入器
		// Replace the response writer
		context.Writer = bufferedWriter

		// 执行下一个中间件
		// Execute the next middleware
		context.Next()

		// 恢复响应写入器
		// Restore the response writer
		context.Writer = bufferedWriter.GetResponseWriter()

		// 重置缓冲区
		// Reset buffer
		bufferedWriter.Reset()

		// 回收缓冲区池对象
		// Recycle buffer pool objects
		if buffer, ok := context.Get(com.RequestBodyBufferKey); ok {
			requestBuffer := buffer.(*bytes.Buffer)
			com.RequestBodyBufferPool.Put(requestBuffer)
			context.Set(com.RequestBodyBufferKey, nil)
		}
		if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
			responseBuffer := buffer.(*bytes.Buffer)
			com.ResponseBodyBufferPool.Put(responseBuffer)
			context.Set(com.ResponseBodyBufferKey, nil)
		}
	}
}
