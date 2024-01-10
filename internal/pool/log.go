package pool

import (
	"sync"

	"github.com/shengyanli1982/orbit/utils/log"
)

// LogEventPool represents a pool of LogEvent objects.
type LogEventPool struct {
	eventPool *sync.Pool
}

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

// Get retrieves a LogEvent from the pool.
func (p *LogEventPool) Get() *log.LogEvent {
	return p.eventPool.Get().(*log.LogEvent)
}

// Put returns a LogEvent to the pool.
func (p *LogEventPool) Put(e *log.LogEvent) {
	if e != nil {
		e.Reset()
		p.eventPool.Put(e)
	}
}
