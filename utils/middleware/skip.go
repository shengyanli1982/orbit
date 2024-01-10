package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// skipPaths contains the paths that should be skipped by the middleware.
var skipPaths = []string{
	com.PromMetricURLPath,  // Prometheus metric URL path
	com.HealthCheckURLPath, // Health check URL path
	com.SwaggerURLPath,     // Swagger UI URL path
	com.PprofURLPath,       // Profiling URL path
}

// SkipResources checks if the request path should be skipped by the middleware.
func SkipResources(c *gin.Context) bool {
	for i := 0; i < len(skipPaths); i++ {
		if strings.HasPrefix(c.Request.URL.Path, skipPaths[i]) {
			return true
		}
	}

	return false
}
