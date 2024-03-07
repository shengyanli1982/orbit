package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// BodyBuffer 是一个中间件，用于缓冲请求和响应的主体。
// BodyBuffer is a middleware that buffers the request and response bodies.
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 从请求主体缓冲池中获取一个缓冲区，并将其设置到当前上下文中
		// Get a buffer from the request body buffer pool and set it to the current context
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)

		// 从响应主体缓冲池中获取一个缓冲区，并将其设置到当前上下文中
		// Get a buffer from the response body buffer pool and set it to the current context
		respBodyBuffer := com.ResponseBodyBufferPool.Get()
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// 创建一个新的缓冲响应写入器，用于写入响应主体
		// Create a new buffer response writer for writing the response body
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)

		// 替换原有的响应写入器为新的缓冲响应写入器
		// Replace the original response writer with the new buffer response writer
		context.Writer = bufferedWriter

		// 执行下一个中间件
		// Execute the next middleware
		context.Next()

		// 恢复原有的响应写入器
		// Restore the original response writer
		context.Writer = bufferedWriter.GetResponseWriter()

		// 重置缓冲响应写入器
		// Reset the buffer response writer
		bufferedWriter.Reset()

		// 如果在当前上下文中找到请求主体缓冲区，则将其放回到请求主体缓冲池中，并在当前上下文中删除它
		// If the request body buffer is found in the current context, put it back to the request body buffer pool and delete it from the current context
		if buffer, ok := context.Get(com.RequestBodyBufferKey); ok {
			requestBuffer := buffer.(*bytes.Buffer)
			com.RequestBodyBufferPool.Put(requestBuffer)
			context.Set(com.RequestBodyBufferKey, nil)
		}

		// 如果在当前上下文中找到响应主体缓冲区，则将其放回到响应主体缓冲池中，并在当前上下文中删除它
		// If the response body buffer is found in the current context, put it back to the response body buffer pool and delete it from the current context
		if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
			responseBuffer := buffer.(*bytes.Buffer)
			com.ResponseBodyBufferPool.Put(responseBuffer)
			context.Set(com.ResponseBodyBufferKey, nil)
		}
	}
}
