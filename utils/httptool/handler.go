package httptool

import (
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

func SkipResources(c *gin.Context) bool {
	if c.Request.URL.Path == com.PromMetricUrlPath ||
		c.Request.URL.Path == com.HttpHealthCheckUrlPath ||
		strings.HasPrefix(c.Request.URL.Path, com.HttpSwaggerUrlPath) ||
		strings.HasPrefix(c.Request.URL.Path, com.HttpPprofUrlPath) {
		return true
	}
	return false
}
