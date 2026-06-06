package middleware

import (
	"github.com/gin-gonic/gin"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// 返回一个 Gin 中间件函数，用于处理请求和响应的缓冲
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, nil)
		originalWriter := context.Writer
		context.Writer = bufferedWriter
		defer func() {
			context.Writer = originalWriter
			bufferedWriter.Reset()
		}()

		context.Next()
	}
}
