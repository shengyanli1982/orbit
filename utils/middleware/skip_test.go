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

		// 测试前缀匹配路径（需满足路径边界：前缀本身或 前缀+"/" 子树）
		// Test prefix match paths (path boundary required: prefix itself or prefix+"/" subtree)
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
		{
			name:     "prefix match - swagger exact prefix",
			path:     com.SwaggerURLPath,
			expected: true,
		},
		{
			name:     "prefix match - pprof exact prefix",
			path:     com.PprofURLPath,
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
			expected: false, // 前缀匹配需路径边界：/docsfake 不属于 /docs 子树
		},
		{
			name:     "no match - documentation is not docs subtree",
			path:     "/documentation",
			expected: false, // 业务路由 /documentation 不得被 /docs 前缀误跳过
		},
		{
			name:     "no match - pprof similar without boundary",
			path:     com.PprofURLPath + "2",
			expected: false,
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
