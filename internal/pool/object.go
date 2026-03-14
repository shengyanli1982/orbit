package pool

import "sync"

// ObjectPool 对 sync.Pool 做泛型封装，避免调用侧散落类型断言。
type ObjectPool[T any] struct {
	pool sync.Pool
}

// NewObjectPool 创建一个泛型对象池。
func NewObjectPool[T any](newFunc func() T) *ObjectPool[T] {
	return &ObjectPool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get 获取对象。
func (p *ObjectPool[T]) Get() T {
	v := p.pool.Get()
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

// Put 归还对象。
func (p *ObjectPool[T]) Put(v T) {
	p.pool.Put(v)
}
