package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// skipPaths 包含了中间件应该跳过的路径
// skipPaths contains the paths that should be skipped by the middleware
var skipPaths = []string{
	// Prometheus 指标 URL 路径
	// Prometheus metric URL path
	com.PromMetricURLPath,

	// 健康检查 URL 路径
	// Health check URL path
	com.HealthCheckURLPath,

	// Swagger UI URL 路径
	// Swagger UI URL path
	com.SwaggerURLPath,

	// 性能分析 URL 路径
	// Profiling URL path
	com.PprofURLPath,
}

// SkipResources 检查请求路径是否应该被中间件跳过
// SkipResources checks if the request path should be skipped by the middleware
func SkipResources(c *gin.Context) bool {
	// 遍历 skipPaths
	// Iterate over skipPaths
	for i := 0; i < len(skipPaths); i++ {
		// 如果请求路径以 skipPaths[i] 为前缀，则返回 true
		// If the request path starts with skipPaths[i], return true
		if strings.HasPrefix(c.Request.URL.Path, skipPaths[i]) {
			return true
		}
	}

	// 如果没有匹配的路径，返回 false
	// If no matching path is found, return false
	return false
}
