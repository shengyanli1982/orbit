package pool

import (
	"sync"

	"github.com/shengyanli1982/orbit/utils/log"
)

// LogEventPool 结构体包含了一个用于管理日志事件对象池的 sync.Pool。
// The LogEventPool struct contains a sync.Pool for managing log event object pooling.
type LogEventPool struct {
	eventPool *sync.Pool // 日志事件对象池 (log event object pool)
}

// NewLogEventPool 函数返回一个新的 LogEventPool 实例。
// The NewLogEventPool function returns a new LogEventPool instance.
func NewLogEventPool() *LogEventPool {
	return &LogEventPool{
		// 创建一个新的 sync.Pool，用于复用 LogEvent 对象
		// Create a new sync.Pool for reusing LogEvent objects
		eventPool: &sync.Pool{
			New: func() interface{} {
				return &log.LogEvent{}
			},
		},
	}
}

// Get 方法从对象池中获取一个日志事件对象。
// The Get method retrieves a log event object from the pool.
func (p *LogEventPool) Get() *log.LogEvent {
	return p.eventPool.Get().(*log.LogEvent)
}

// Put 方法将日志事件对象放回对象池中。
// The Put method returns a log event object to the pool.
func (p *LogEventPool) Put(e *log.LogEvent) {
	// 检查输入对象是否为空
	// Check if the input object is nil
	if e != nil {
		e.Reset()
		p.eventPool.Put(e)
	}
}
