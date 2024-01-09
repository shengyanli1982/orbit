package pool

import (
	"bytes"
	"sync"
)

// DefaultBufferSize is the default size of the buffer.
const DefaultBufferSize = 2048

// BufferPool represents a pool of buffers.
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool with the specified buffer size.
// If the bufferSize is less than or equal to 0, the default buffer size is used.
func NewBufferPool(bufferSize uint32) *BufferPool {
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
	}
}

// Get retrieves a buffer from the pool.
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put returns a buffer to the pool.
// If the buffer is not nil, it is reset before being put back into the pool.
func (p *BufferPool) Put(buffer *bytes.Buffer) {
	if buffer != nil {
		buffer.Reset()
		p.pool.Put(buffer)
	}
}
