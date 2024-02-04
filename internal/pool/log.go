package pool

import (
	"sync"

	"github.com/shengyanli1982/orbit/utils/log"
)

// LogEventPool 表示 LogEvent 对象的池。
// LogEventPool represents a pool of LogEvent objects.
type LogEventPool struct {
	eventPool *sync.Pool
}

// NewLogEventPool 创建一个新的 LogEventPool。
// NewLogEventPool creates a new LogEventPool.
func NewLogEventPool() *LogEventPool {
	return &LogEventPool{
		eventPool: &sync.Pool{
			New: func() interface{} {
				return &log.LogEvent{}
			},
		},
	}
}

// Get 从池中获取一个 LogEvent。
// Get retrieves a LogEvent from the pool.
func (p *LogEventPool) Get() *log.LogEvent {
	return p.eventPool.Get().(*log.LogEvent)
}

// Put 将 LogEvent 放回到池中。
// Put returns a LogEvent to the pool.
func (p *LogEventPool) Put(e *log.LogEvent) {
	if e != nil {
		e.Reset()
		p.eventPool.Put(e)
	}
}
