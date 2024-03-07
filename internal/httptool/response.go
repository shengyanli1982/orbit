package httptool

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// ResponseBodyWriter 结构体包含一个 gin.ResponseWriter 接口和一个 bytes.Buffer 实例。
// The ResponseBodyWriter struct contains a gin.ResponseWriter interface and a bytes.Buffer instance.
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

// Write 是 ResponseBodyWriter 的一个方法，它将字节切片 b 写入连接。
// Write is a method of ResponseBodyWriter that writes the byte slice b to the connection.
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	// 首先，尝试将字节切片 b 写入缓冲区。如果出错，返回写入的字节数和错误。
	// First, try to write the byte slice b to the buffer. If there is an error, return the number of bytes written and the error.
	if count, err := w.buffer.Write(b); err != nil {
		return count, err
	}

	// 然后，将字节切片 b 写入 ResponseWriter。返回写入的字节数和错误（如果有）。
	// Then, write the byte slice b to the ResponseWriter. Return the number of bytes written and any error.
	return w.ResponseWriter.Write(b)
}

// WriteString 是 ResponseBodyWriter 的一个方法，它将字符串 s 写入 w，w 接受一个字节切片。
// WriteString is a method of ResponseBodyWriter that writes the string s to w, which accepts a slice of bytes.
func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	// 首先，尝试将字符串 s 写入缓冲区。如果出错，返回写入的字节数和错误。
	// First, try to write the string s to the buffer. If there is an error, return the number of bytes written and the error.
	if count, err := w.buffer.WriteString(s); err != nil {
		return count, err
	}

	// 然后，将字符串 s 写入 ResponseWriter。返回写入的字节数和错误（如果有）。
	// Then, write the string s to the ResponseWriter. Return the number of bytes written and any error.
	return w.ResponseWriter.WriteString(s)
}

// Reset 重置缓冲区。
// Reset resets the buffer.
func (w *ResponseBodyWriter) Reset() {
	w.buffer = nil
}

// GetBuffer 返回缓冲区。
// GetBuffer returns the buffer.
func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer
}

// GetResponseWriter 返回 ResponseWriter。
// GetResponseWriter returns the ResponseWriter.
func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter
}
