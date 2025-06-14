package httptool

import (
	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"go.uber.org/zap"
)

// 从上下文中返回日志记录器
func GetLoggerFromContext(context *gin.Context) *zap.SugaredLogger {
	if obj, ok := context.Get(com.RequestLoggerKey); ok {
		return obj.(*zap.SugaredLogger)
	}

	return com.DefaultSugeredLogger
}
