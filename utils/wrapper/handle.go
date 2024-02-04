package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHandlerFuncToGin 包装 http.HandlerFunc 到 gin.HandlerFunc。
// 它通过使用提供的处理程序来服务 HTTP 请求，将输入的 http.HandlerFunc 转换为 gin.HandlerFunc。
// WrapHandlerFuncToGin wraps an http.HandlerFunc to a gin.HandlerFunc.
// It converts the input http.HandlerFunc to a gin.HandlerFunc by serving the HTTP request using the provided handler.
func WrapHandlerFuncToGin(handler http.HandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	}
}

// WrapHandlerToGin 包装 http.Handler 到 gin.HandlerFunc。
// 它通过使用提供的处理程序来服务 HTTP 请求，将输入的 http.Handler 转换为 gin.HandlerFunc。
// WrapHandlerToGin wraps an http.Handler to a gin.HandlerFunc.
// It converts the input http.Handler to a gin.HandlerFunc by serving the HTTP request using the provided handler.
func WrapHandlerToGin(handler http.Handler) gin.HandlerFunc {
	return func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	}
}
