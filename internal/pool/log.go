package pool

import "sync"

type LogEvent struct {
	Message        string `json:"message,omitempty" yaml:"message,omitempty"`
	ID             string `json:"id,omitempty" yaml:"id,omitempty"`
	IP             string `json:"ip,omitempty" yaml:"ip,omitempty"`
	EndPoint       string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Path           string `json:"path,omitempty" yaml:"path,omitempty"`
	Method         string `json:"method,omitempty" yaml:"method,omitempty"`
	Code           int    `json:"statuscode,omitempty" yaml:"statucode,omitempty"`
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
	bp *sync.Pool
}

func NewLogEventPool() *LogEventPool {
	return &LogEventPool{
		bp: &sync.Pool{
			New: func() interface{} {
				return &LogEvent{}
			},
		},
	}
}

func (p *LogEventPool) Get() *LogEvent {
	return p.bp.Get().(*LogEvent)
}

func (p *LogEventPool) Put(e *LogEvent) {
	if e != nil {
		e.Reset()
		p.bp.Put(e)
	}
}
