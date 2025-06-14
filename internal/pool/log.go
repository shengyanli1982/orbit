package pool

import (
	"sync"

	"github.com/shengyanli1982/orbit/utils/log"
)

// 包含了一个用于管理日志事件对象池的 sync.Pool
type LogEventPool struct {
	eventPool *sync.Pool // 日志事件对象池
}

// 返回一个新的 LogEventPool 实例
func NewLogEventPool() *LogEventPool {
	return &LogEventPool{
		// 创建一个新的 sync.Pool，用于复用 LogEvent 对象
		eventPool: &sync.Pool{
			New: func() interface{} {
				return &log.LogEvent{}
			},
		},
	}
}

// 从对象池中获取一个日志事件对象
func (p *LogEventPool) Get() *log.LogEvent {
	return p.eventPool.Get().(*log.LogEvent)
}

// 将日志事件对象放回对象池中
func (p *LogEventPool) Put(e *log.LogEvent) {
	// 检查输入对象是否为空
	if e != nil {
		e.Reset()
		p.eventPool.Put(e)
	}
}
