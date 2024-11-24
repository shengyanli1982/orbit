package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHandlerFuncToGin 函数将 http.HandlerFunc 包装为 gin.HandlerFunc。
// 它通过使用提供的处理器来服务 HTTP 请求，将输入的 http.HandlerFunc 转换为 gin.HandlerFunc。
// The WrapHandlerFuncToGin function wraps an http.HandlerFunc to a gin.HandlerFunc.
// It converts the input http.HandlerFunc to a gin.HandlerFunc by serving the HTTP request using the provided handler.
func WrapHandlerFuncToGin(handler http.HandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	}
}

// WrapHandlerToGin 函数将 http.Handler 包装为 gin.HandlerFunc。
// 它通过使用提供的处理器来服务 HTTP 请求，将输入的 http.Handler 转换为 gin.HandlerFunc。
// The WrapHandlerToGin function wraps an http.Handler to a gin.HandlerFunc.
// It converts the input http.Handler to a gin.HandlerFunc by serving the HTTP request using the provided handler.
func WrapHandlerToGin(handler http.Handler) gin.HandlerFunc {
	return func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	}
}
