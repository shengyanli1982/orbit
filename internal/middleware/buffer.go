package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	omid "github.com/shengyanli1982/orbit/utils/middleware"
)

func BodyBuffer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip resources that do not need to be recorded
		if omid.SkipResources(c) {
			c.Next()
			return
		}

		// Set request and response body buffers
		c.Set(com.RequestBodyBufferKey, com.RequestBodyBufferPool.Get())
		c.Set(com.ResponseBodyBufferKey, com.ResponseBodyBufferPool.Get())

		// Execute the next middleware
		c.Next()

		// Recycle buffer pool objects
		if requestBodyBuffer, ok := c.Get(com.RequestBodyBufferKey); ok {
			requestBuffer := requestBodyBuffer.(*bytes.Buffer)
			com.RequestBodyBufferPool.Put(requestBuffer)
		}
		if responseBodyBuffer, ok := c.Get(com.ResponseBodyBufferKey); ok {
			responseBuffer := responseBodyBuffer.(*bytes.Buffer)
			com.ResponseBodyBufferPool.Put(responseBuffer)
		}
	}
}
