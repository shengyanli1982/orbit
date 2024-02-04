package httptool

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// ResponseBodyWriter 是 gin.ResponseWriter 的封装，用于捕获响应体。
// ResponseBodyWriter is a wrapper around gin.ResponseWriter that captures the response body.
type ResponseBodyWriter struct {
	// gin.ResponseWriter 是一个封装了 http.ResponseWriter 接口的接口。
	// gin.ResponseWriter is an interface that encapsulates the http.ResponseWriter interface.
	gin.ResponseWriter

	// buffer 是一个 bytes.Buffer 实例。
	// buffer is a bytes.Buffer instance.
	buffer *bytes.Buffer
}

// NewResponseBodyWriter 返回一个新的 ResponseBodyWriter 实例。
// NewResponseBodyWriter returns a new ResponseBodyWriter instance.
func NewResponseBodyWriter(w gin.ResponseWriter, buf *bytes.Buffer) *ResponseBodyWriter {
	return &ResponseBodyWriter{
		ResponseWriter: w,
		buffer:         buf,
	}
}

// Write 写入数据到连接，同时保存一份内容到 buffer。
// Write writes data to the connection and saves a copy of the content to the buffer.
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	if count, err := w.buffer.Write(b); err != nil {
		return count, err
	}
	return w.ResponseWriter.Write(b)
}

// WriteString 将字符串写入到连接，同时保存一份内容到 buffer。
// WriteString writes a string to the connection and saves a copy of the content to the buffer.
func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	if count, err := w.buffer.WriteString(s); err != nil {
		return count, err
	}
	return w.ResponseWriter.WriteString(s)
}

// Reset 重置 buffer。
// Reset resets the buffer.
func (w *ResponseBodyWriter) Reset() {
	w.buffer = nil
}

// GetBuffer 返回 buffer。
// GetBuffer returns the buffer.
func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer
}

// GetResponseWriter 返回 ResponseWriter。
// GetResponseWriter returns the ResponseWriter.
func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter
}
