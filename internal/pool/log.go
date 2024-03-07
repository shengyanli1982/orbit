package pool

import (
	"sync"

	"github.com/shengyanli1982/orbit/utils/log"
)

// LogEventPool 表示 LogEvent 对象的池。
// LogEventPool represents a pool of LogEvent objects.
type LogEventPool struct {
	eventPool *sync.Pool // 用于存储 LogEvent 对象的同步池 (A sync pool for storing LogEvent objects)
}

// NewLogEventPool 创建一个新的 LogEventPool。
// NewLogEventPool creates a new LogEventPool.
func NewLogEventPool() *LogEventPool {
	return &LogEventPool{
		eventPool: &sync.Pool{
			// 当池中没有可用对象时，New 函数将创建一个新的 LogEvent 对象。
			// The New function creates a new LogEvent object when there are no available objects in the pool.
			New: func() interface{} {
				return &log.LogEvent{}
			},
		},
	}
}

// Get 从池中检索一个 LogEvent。
// Get retrieves a LogEvent from the pool.
func (p *LogEventPool) Get() *log.LogEvent {
	// 从 eventPool 中获取一个 LogEvent 对象，并将其转换为正确的类型。
	// Get a LogEvent object from the eventPool and cast it to the correct type.
	return p.eventPool.Get().(*log.LogEvent)
}

// Put 将一个 LogEvent 返回到池中。
// Put returns a LogEvent to the pool.
func (p *LogEventPool) Put(e *log.LogEvent) {
	if e != nil {
		// 如果 LogEvent 对象不为空，则重置对象并将其放回到池中。
		// If the LogEvent object is not nil, reset the object and put it back into the pool.
		e.Reset()
		p.eventPool.Put(e)
	}
}
