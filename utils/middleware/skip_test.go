package middleware

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSkipResources(t *testing.T) {
	c := gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				Path: "/some/path",
			},
		},
	}

	// Test when the request path matches one of the skipPaths
	for _, skipPath := range skipPaths {
		c.Request.URL.Path = skipPath
		assert.True(t, SkipResources(&c), "Expected SkipResources to return true for path %s, but got false", skipPath)
	}

	// Test when the request path does not match any skipPaths
	c.Request.URL.Path = "/some/other/path"
	assert.False(t, SkipResources(&c), "Expected SkipResources to return false for path %s, but got true", c.Request.URL.Path)
}
