package httptool

import (
	"bytes"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockResponseWriter struct {
	gin.ResponseWriter
	written []byte
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	p := []byte(s)
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockResponseWriter) Flush() {}

func TestResponseBodyWriter_Write(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	data := []byte("test data")
	n, err := w.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), w.Size())
	assert.Equal(t, data, buf.Bytes())
}

func TestResponseBodyWriter_WriteString(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	data := "test data"
	n, err := w.WriteString(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), w.Size())
	assert.Equal(t, []byte(data), buf.Bytes())
}

func TestResponseBodyWriter_Flush(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	testData := "test data"
	_, err := w.WriteString(testData)
	assert.NoError(t, err)

	assert.Equal(t, testData, buf.String())
	assert.Equal(t, testData, string(mock.written))

	mock.written = mock.written[:0]

	w.Flush()

	assert.Equal(t, testData, string(mock.written))
	assert.Equal(t, 0, buf.Len())
}

func BenchmarkResponseBodyWriter_Write(b *testing.B) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func BenchmarkResponseBodyWriter_WriteString(b *testing.B) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)
	data := "test data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteString(data)
	}
}
