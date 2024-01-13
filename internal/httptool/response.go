package httptool

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type ResponseBodyWriter struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func NewResponseBodyWriter(w gin.ResponseWriter, buf *bytes.Buffer) *ResponseBodyWriter {
	return &ResponseBodyWriter{
		ResponseWriter: w,
		buffer:         buf,
	}
}

func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	if count, err := w.buffer.Write(b); err != nil {
		return count, err
	}
	return w.ResponseWriter.Write(b)
}

func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	if count, err := w.buffer.WriteString(s); err != nil {
		return count, err
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *ResponseBodyWriter) Reset() {
	w.buffer = nil
}

func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer
}

func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter
}
