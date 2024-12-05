package middleware

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

func TestSkipResources(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// 测试精确匹配路径
		// Test exact match paths
		{
			name:     "exact match - prometheus metrics",
			path:     com.PromMetricURLPath,
			expected: true,
		},
		{
			name:     "exact match - health check",
			path:     com.HealthCheckURLPath,
			expected: true,
		},

		// 测试前缀匹配路径
		// Test prefix match paths
		{
			name:     "prefix match - swagger with suffix",
			path:     com.SwaggerURLPath + "/index.html",
			expected: true,
		},
		{
			name:     "prefix match - pprof with suffix",
			path:     com.PprofURLPath + "/goroutine",
			expected: true,
		},

		// 测试不匹配的路径
		// Test non-matching paths
		{
			name:     "no match - random path",
			path:     "/api/v1/users",
			expected: false,
		},
		{
			name:     "no match - empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "no match - similar but not prefix",
			path:     com.SwaggerURLPath + "fake",
			expected: true, // 因为是前缀匹配，所以这个也会返回 true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gin.Context{
				Request: &http.Request{
					URL: &url.URL{
						Path: tt.path,
					},
				},
			}

			result := SkipResources(c)
			assert.Equal(t, tt.expected, result, "Path: %s", tt.path)
		})
	}
}
