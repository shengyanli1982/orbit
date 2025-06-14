package log

// LogEvent 结构体用于记录日志事件
type LogEvent struct {
	// 日志消息
	Message string `json:"message,omitempty" yaml:"message,omitempty"`

	// 事件的唯一标识符
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// 发起请求的IP地址
	IP string `json:"ip,omitempty" yaml:"ip,omitempty"`

	// 请求的终端点
	EndPoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// 请求的路径
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// 请求的HTTP方法
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// 响应的HTTP状态码
	Code int `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`

	// 请求的状态
	Status string `json:"status,omitempty" yaml:"status,omitempty"`

	// 请求的延迟时间
	Latency string `json:"latency,omitempty" yaml:"latency,omitempty"`

	// 发起请求的用户代理
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`

	// 请求的内容类型
	ReqContentType string `json:"reqContentType,omitempty" yaml:"reqContentType,omitempty"`

	// 请求的查询参数
	ReqQuery string `json:"query,omitempty" yaml:"query,omitempty"`

	// 请求的主体内容
	ReqBody string `json:"reqBody,omitempty" yaml:"reqBody,omitempty"`

	// 请求中的任何错误
	Error error `json:"error,omitempty" yaml:"error,omitempty"`

	// 错误的堆栈跟踪
	ErrorStack string `json:"errorStack,omitempty" yaml:"errorStack,omitempty"`
}

// Reset 方法用于重置 LogEvent 结构体的所有字段
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
