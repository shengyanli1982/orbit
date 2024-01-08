package pool

import (
	"bytes"
	"sync"
)

const DefaultBufferSize = 2048

type BufferPool struct {
	bp sync.Pool
}

func NewBufferPool(size uint32) *BufferPool {
	if size <= 0 {
		size = DefaultBufferSize
	}
	return &BufferPool{
		bp: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, size))
			},
		},
	}
}

func (p *BufferPool) Get() *bytes.Buffer {
	return p.bp.Get().(*bytes.Buffer)
}

func (p *BufferPool) Put(b *bytes.Buffer) {
	if b != nil {
		b.Reset()
		p.bp.Put(b)
	}
}
