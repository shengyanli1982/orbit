package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	ihttptool "github.com/shengyanli1982/orbit/internal/httptool"
)

// BodyBuffer is a middleware that buffers the request and response bodies.
func BodyBuffer() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Set request and response body buffers
		reqBodyBuffer := com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		respBodyBuffer := com.ResponseBodyBufferPool.Get()
		context.Set(com.ResponseBodyBufferKey, respBodyBuffer)

		// Create a new buffer response writer
		bufferedWriter := ihttptool.NewResponseBodyWriter(context.Writer, respBodyBuffer)
		// Replace the response writer
		context.Writer = bufferedWriter

		// Execute the next middleware
		context.Next()

		// Restore the response writer
		context.Writer = bufferedWriter.GetResponseWriter()
		// Reset buffer
		bufferedWriter.Reset()

		// Recycle buffer pool objects
		if buffer, ok := context.Get(com.RequestBodyBufferKey); ok {
			requestBuffer := buffer.(*bytes.Buffer)
			com.RequestBodyBufferPool.Put(requestBuffer)
			context.Set(com.RequestBodyBufferKey, nil)
		}
		if buffer, ok := context.Get(com.ResponseBodyBufferKey); ok {
			responseBuffer := buffer.(*bytes.Buffer)
			com.ResponseBodyBufferPool.Put(responseBuffer)
			context.Set(com.ResponseBodyBufferKey, nil)
		}
	}
}
