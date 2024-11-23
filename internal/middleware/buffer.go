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
		// 获取缓冲区
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		respBodyBuffer := com.ResponseBodyBufferPool.Get()

		// 设置缓冲区到上下文
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// 创建并设置缓冲响应写入器
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)
		originalWriter := context.Writer
		context.Writer = bufferedWriter

		// 执行下一个中间件
		context.Next()

		// 恢复原始写入器并重置缓冲写入器
		context.Writer = originalWriter
		bufferedWriter.Reset()

		// 清理并返回请求缓冲区
		if reqBuffer, ok := context.Get(com.RequestBodyBufferKey); ok {
			if buf, ok := reqBuffer.(*bytes.Buffer); ok {
				buf.Reset()
				com.RequestBodyBufferPool.Put(buf)
			}
			context.Set(com.RequestBodyBufferKey, nil)
		}

		// 清理并返回响应缓冲区
		if respBuffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
			if buf, ok := respBuffer.(*bytes.Buffer); ok {
				buf.Reset()
				com.ResponseBodyBufferPool.Put(buf)
			}
			context.Set(com.ResponseBodyBufferKey, nil)
		}
	}
}
