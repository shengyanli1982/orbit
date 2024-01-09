package pool

import (
	"bytes"
	"sync"
)

const DefaultBufferSize = 2048

type BufferPool struct {
	pool sync.Pool
}

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

func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *BufferPool) Put(buffer *bytes.Buffer) {
	if buffer != nil {
		buffer.Reset()
		p.pool.Put(buffer)
	}
}
