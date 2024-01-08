package httptool

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

var skipPaths = []string{
	com.PromMetricUrlPath,
	com.HttpHealthCheckUrlPath,
	com.HttpSwaggerUrlPath,
	com.HttpPprofUrlPath,
}

func SkipResources(c *gin.Context) bool {
	for i := 0; i < len(skipPaths); i++ {
		if strings.HasPrefix(c.Request.URL.Path, skipPaths[i]) {
			return true
		}
	}

	return false
}
