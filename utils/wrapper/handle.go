package wrapper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHandlerFuncToGin wraps an http.HandlerFunc to a gin.HandlerFunc.
// It converts the input http.HandlerFunc to a gin.HandlerFunc by serving the HTTP request using the provided handler.
func WrapHandlerFuncToGin(handler http.HandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	}
}
