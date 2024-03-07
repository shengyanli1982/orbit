package httptool

import (
	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"go.uber.org/zap"
)

// GetLoggerFromContext 函数从上下文中返回日志记录器
// The GetLoggerFromContext function returns the logger from the context
func GetLoggerFromContext(context *gin.Context) *zap.SugaredLogger {
	// 如果上下文中存在 com.RequestLoggerKey，那么返回对应的日志记录器
	// If the com.RequestLoggerKey exists in the context, then return the corresponding logger
	if obj, ok := context.Get(com.RequestLoggerKey); ok {
		// 将 obj 转换为 *zap.SugaredLogger 类型并返回
		// Convert obj to *zap.SugaredLogger type and return
		return obj.(*zap.SugaredLogger)
	}
	// 如果上下文中不存在 com.RequestLoggerKey，那么返回默认的日志记录器
	// If the com.RequestLoggerKey does not exist in the context, then return the default logger
	return com.DefaultSugeredLogger
}
