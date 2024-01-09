package pool

import (
	"sync"
)

type LogEvent struct {
	Message        string `json:"message,omitempty" yaml:"message,omitempty"`
	ID             string `json:"id,omitempty" yaml:"id,omitempty"`
	IP             string `json:"ip,omitempty" yaml:"ip,omitempty"`
	EndPoint       string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Path           string `json:"path,omitempty" yaml:"path,omitempty"`
	Method         string `json:"method,omitempty" yaml:"method,omitempty"`
	Code           int    `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`
	Status         string `json:"status,omitempty" yaml:"status,omitempty"`
	Latency        string `json:"latency,omitempty" yaml:"latency,omitempty"`
	Agent          string `json:"agent,omitempty" yaml:"agent,omitempty"`
	ReqContentType string `json:"reqContentType,omitempty" yaml:"reqContentType,omitempty"`
	ReqQuery       string `json:"query,omitempty" yaml:"query,omitempty"`
	ReqBody        string `json:"reqBody,omitempty" yaml:"reqBody,omitempty"`
	Error          any    `json:"error,omitempty" yaml:"error,omitempty"`
	ErrorStack     string `json:"errorStack,omitempty" yaml:"errorStack,omitempty"`
}

func (e *LogEvent) Reset() {
	e.Message = ""
	e.ID = ""
	e.IP = ""
	e.EndPoint = ""
	e.Path = ""
	e.Method = ""
	e.Code = 0
	e.Status = ""
	e.Latency = ""
	e.Agent = ""
	e.ReqContentType = ""
	e.ReqQuery = ""
	e.ReqBody = ""
	e.Error = nil
	e.ErrorStack = ""
}

type LogEventPool struct {
	eventPool *sync.Pool
}

func NewLogEventPool() *LogEventPool {
	return &LogEventPool{
		eventPool: &sync.Pool{
			New: func() interface{} {
				return &LogEvent{}
			},
		},
	}
}

func (p *LogEventPool) Get() *LogEvent {
	return p.eventPool.Get().(*LogEvent)
}

func (p *LogEventPool) Put(e *LogEvent) {
	if e != nil {
		e.Reset()
		p.eventPool.Put(e)
	}
}
