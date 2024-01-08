package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlerFuncToGin(h http.HandlerFunc) gin.HandlerFunc {
	// 返回一个匿名函数作为 gin.HandlerFunc
	return func(c *gin.Context) {
		// 调用输入的 http.Handler 的 ServeHTTP 方法
		// 并将 gin.Context.Writer 和 gin.Context.Request 传递作为参数
		h.ServeHTTP(c.Writer, c.Request)
	}
}
