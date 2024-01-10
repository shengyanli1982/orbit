package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// BodyBuffer is a middleware that buffers the request and response bodies.
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Set request and response body buffers
		context.Set(com.RequestBodyBufferKey, com.RequestBodyBufferPool.Get())
		context.Set(com.ResponseBodyBufferKey, com.ResponseBodyBufferPool.Get())

		// Execute the next middleware
		context.Next()

		// Recycle buffer pool objects
		if requestBodyBuffer, ok := context.Get(com.RequestBodyBufferKey); ok {
			requestBuffer := requestBodyBuffer.(*bytes.Buffer)
			com.RequestBodyBufferPool.Put(requestBuffer)
		}
		if responseBodyBuffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
			responseBuffer := responseBodyBuffer.(*bytes.Buffer)
			com.ResponseBodyBufferPool.Put(responseBuffer)
		}
	}
}
