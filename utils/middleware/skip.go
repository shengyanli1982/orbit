package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// 跳过路径的配置结构
type skipConfig struct {
	exact       map[string]struct{} // 精确匹配的路径
	prefix      []string            // 前缀匹配的路径
	prefixSlash []string            // 预计算的 prefix+"/"，与 prefix 等长同序，消除热路径重复拼接
}

// 需要跳过中间件处理的 URL 路径配置
var skipPaths = newSkipConfig()

// 构造跳过路径配置：prefix+"/" 仅在包初始化时拼接一次
func newSkipConfig() *skipConfig {
	cfg := &skipConfig{
		exact: map[string]struct{}{
			com.PromMetricURLPath:  {}, // Prometheus 指标路径
			com.HealthCheckURLPath: {}, // 健康检查路径
		},
		prefix: []string{
			com.SwaggerURLPath, // Swagger API 文档路径
			com.PprofURLPath,   // pprof 性能分析路径
		},
	}
	cfg.prefixSlash = make([]string, len(cfg.prefix))
	for i, prefix := range cfg.prefix {
		cfg.prefixSlash[i] = prefix + "/"
	}
	return cfg
}

// 检查当前请求是否应该跳过中间件处理
func SkipResources(c *gin.Context) bool {
	path := c.Request.URL.Path

	// 首先尝试精确匹配，这个速度最快
	if _, exists := skipPaths.exact[path]; exists {
		return true
	}

	// 如果精确匹配失败，尝试前缀匹配（带路径边界：仅匹配前缀本身或 前缀+"/" 子树，
	// 避免 /docsfake 之类的业务路由被 /docs 前缀误跳过）
	for i, prefix := range skipPaths.prefix {
		if path == prefix || strings.HasPrefix(path, skipPaths.prefixSlash[i]) {
			return true
		}
	}

	return false
}
