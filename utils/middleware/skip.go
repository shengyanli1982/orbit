package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// skipPaths 包含了需要跳过中间件处理的 URL 路径。
// skipPaths contains URL paths that should bypass middleware processing.
var skipPaths = []string{
	com.PromMetricURLPath,  // Prometheus 指标路径 (Prometheus metrics path)
	com.HealthCheckURLPath, // 健康检查路径 (Health check path)
	com.SwaggerURLPath,     // Swagger API 文档路径 (Swagger API documentation path)
	com.PprofURLPath,       // pprof 性能分析路径 (pprof profiling path)
}

// SkipResources 函数检查当前请求是否应该跳过中间件处理。
// 如果请求路径以 skipPaths 中的任何路径为前缀，则返回 true。
// The SkipResources function checks if the current request should bypass middleware processing.
// Returns true if the request path starts with any path in skipPaths.
func SkipResources(c *gin.Context) bool {
	// 遍历所有需要跳过的路径
	// Iterate through all paths that should be skipped
	for i := 0; i < len(skipPaths); i++ {
		// 检查当前请求路径是否以跳过路径为前缀
		// Check if the current request path starts with the skip path
		if strings.HasPrefix(c.Request.URL.Path, skipPaths[i]) {
			return true
		}
	}

	// 返回 false 表示不需要跳过
	// Return false to indicate that the path should not be skipped
	return false
}
