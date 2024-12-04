package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHandlerFuncToGin 将 http.HandlerFunc 包装为 gin.HandlerFunc
// 直接使用原始处理器，避免额外的函数调用开销
func WrapHandlerFuncToGin(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(c.Writer, c.Request)
	}
}

// WrapHandlerToGin 将 http.Handler 包装为 gin.HandlerFunc
// 直接使用原始处理器，避免额外的函数调用开销
func WrapHandlerToGin(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
