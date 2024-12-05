package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHandlerFuncToGin 将 http.HandlerFunc 包装为 gin.HandlerFunc
// 直接使用原始处理器，避免额外的函数调用开销
func WrapHandlerFuncToGin(fn http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(c.Writer, c.Request)
	}
}

// WrapHandlerToGin 将 http.Handler 包装为 gin.HandlerFunc
// 直接使用原始处理器，避免额外的函数调用开销
func WrapHandlerToGin(fn http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn.ServeHTTP(c.Writer, c.Request)
	}
}
