package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// 跳过路径的配置结构
type skipConfig struct {
	exact  map[string]struct{} // 精确匹配的路径
	prefix []string            // 前缀匹配的路径
}

// 需要跳过中间件处理的 URL 路径配置
var skipPaths = &skipConfig{
	exact: map[string]struct{}{
		com.PromMetricURLPath:  {}, // Prometheus 指标路径
		com.HealthCheckURLPath: {}, // 健康检查路径
	},
	prefix: []string{
		com.SwaggerURLPath, // Swagger API 文档路径
		com.PprofURLPath,   // pprof 性能分析路径
	},
}

// 检查当前请求是否应该跳过中间件处理
func SkipResources(c *gin.Context) bool {
	path := c.Request.URL.Path

	// 首先尝试精确匹配，这个速度最快
	if _, exists := skipPaths.exact[path]; exists {
		return true
	}

	// 如果精确匹配失败，尝试前缀匹配
	for _, prefix := range skipPaths.prefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
