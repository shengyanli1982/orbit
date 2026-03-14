package middleware

import (
	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// 返回一个 Gin 中间件函数，用于处理请求和响应的缓冲
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 从缓冲池中获取缓冲区
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		respBodyBuffer := com.ResponseBodyBufferPool.Get()

		// 预先清空缓冲区，确保没有残留数据
		reqBodyBuffer.Reset()
		respBodyBuffer.Reset()

		// 将缓冲区存储在上下文中
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// 创建并设置响应写入器
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)
		originalWriter := context.Writer
		context.Writer = bufferedWriter
		defer func() {
			// 恢复原始写入器，并归还对象池资源。
			// 注意：respBodyBuffer 由 bufferedWriter.Reset() 归还，避免重复 Put 导致并发复用。
			context.Writer = originalWriter
			bufferedWriter.Reset()
			reqBodyBuffer.Reset()
			com.RequestBodyBufferPool.Put(reqBodyBuffer)
		}()

		// 执行后续的中间件和处理函数
		context.Next()
	}
}
