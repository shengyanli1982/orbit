package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WrapHandlerFuncToGin(handler http.HandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	}
}
