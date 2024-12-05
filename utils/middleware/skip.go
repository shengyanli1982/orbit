package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// skipConfig 定义了跳过路径的配置结构
// skipConfig defines the configuration structure for skip paths
type skipConfig struct {
	exact  map[string]struct{} // 精确匹配的路径 paths for exact match
	prefix []string            // 前缀匹配的路径 paths for prefix match
}

// skipPaths 包含了需要跳过中间件处理的 URL 路径配置
// skipPaths contains URL paths configuration that should bypass middleware processing
var skipPaths = &skipConfig{
	exact: map[string]struct{}{
		com.PromMetricURLPath:  {}, // Prometheus 指标路径 (Prometheus metrics path)
		com.HealthCheckURLPath: {}, // 健康检查路径 (Health check path)
	},
	prefix: []string{
		com.SwaggerURLPath, // Swagger API 文档路径 (Swagger API documentation path)
		com.PprofURLPath,   // pprof 性能分析路径 (pprof profiling path)
	},
}

// SkipResources 函数检查当前请求是否应该跳过中间件处理
// The SkipResources function checks if the current request should bypass middleware processing
func SkipResources(c *gin.Context) bool {
	path := c.Request.URL.Path

	// 首先尝试精确匹配，这个速度最快
	// Try exact match first, this is the fastest
	if _, exists := skipPaths.exact[path]; exists {
		return true
	}

	// 如果精确匹配失败，尝试前缀匹配
	// If exact match fails, try prefix match
	for _, prefix := range skipPaths.prefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
