package httptool

import (
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	com "github.com/shengyanli1982/orbit/common"
)

// GetLoggerFromContext 从 gin.Context 中获取日志记录器
// 如果上下文中存在日志记录器则返回，否则返回默认日志记录器
func GetLoggerFromContext(context *gin.Context) *logr.Logger {
	if context == nil {
		return &com.DefaultLogrLogger
	}

	if obj, ok := context.Get(com.RequestLoggerKey); ok {
		if logger, ok := obj.(*logr.Logger); ok && logger != nil {
			return logger
		}
	}

	return &com.DefaultLogrLogger
}
