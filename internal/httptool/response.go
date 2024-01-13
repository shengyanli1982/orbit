package httptool

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type ResponseBodyWriter struct {
	// gin.ResponseWriter is an interface that encapsulates the http.ResponseWriter interface.
	gin.ResponseWriter
	// buffer is a bytes.Buffer instance.
	buffer *bytes.Buffer
}

// NewResponseBodyWriter returns a new ResponseBodyWriter instance.
func NewResponseBodyWriter(w gin.ResponseWriter, buf *bytes.Buffer) *ResponseBodyWriter {
	return &ResponseBodyWriter{
		ResponseWriter: w,
		buffer:         buf,
	}
}

// Write writes the data to the connection.
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	if count, err := w.buffer.Write(b); err != nil {
		return count, err
	}
	return w.ResponseWriter.Write(b)
}

// WriteString writes the contents of the string s to w, which accepts a slice of bytes.
func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	if count, err := w.buffer.WriteString(s); err != nil {
		return count, err
	}
	return w.ResponseWriter.WriteString(s)
}

// Reset resets the buffer.
func (w *ResponseBodyWriter) Reset() {
	w.buffer = nil
}

// GetBuffer returns the buffer.
func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer
}

// GetResponseWriter returns the ResponseWriter.
func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter
}
