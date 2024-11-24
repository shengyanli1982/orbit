package log

// LogEvent 结构体用于记录日志事件
// The LogEvent struct is used to log events
type LogEvent struct {
	// Message 字段表示日志消息
	// The Message field represents the log message
	Message string `json:"message,omitempty" yaml:"message,omitempty"`

	// ID 字段表示事件的唯一标识符
	// The ID field represents the unique identifier of the event
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// IP 字段表示发起请求的IP地址
	// The IP field represents the IP address of the request initiator
	IP string `json:"ip,omitempty" yaml:"ip,omitempty"`

	// EndPoint 字段表示请求的终端点
	// The EndPoint field represents the endpoint of the request
	EndPoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Path 字段表示请求的路径
	// The Path field represents the path of the request
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Method 字段表示请求的HTTP方法
	// The Method field represents the HTTP method of the request
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Code 字段表示响应的HTTP状态码
	// The Code field represents the HTTP status code of the response
	Code int `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`

	// Status 字段表示请求的状态
	// The Status field represents the status of the request
	Status string `json:"status,omitempty" yaml:"status,omitempty"`

	// Latency 字段表示请求的延迟时间
	// The Latency field represents the latency of the request
	Latency string `json:"latency,omitempty" yaml:"latency,omitempty"`

	// Agent 字段表示发起请求的用户代理
	// The Agent field represents the user agent of the request initiator
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`

	// ReqContentType 字段表示请求的内容类型
	// The ReqContentType field represents the content type of the request
	ReqContentType string `json:"reqContentType,omitempty" yaml:"reqContentType,omitempty"`

	// ReqQuery 字段表示请求的查询参数
	// The ReqQuery field represents the query parameters of the request
	ReqQuery string `json:"query,omitempty" yaml:"query,omitempty"`

	// ReqBody 字段表示请求的主体内容
	// The ReqBody field represents the body of the request
	ReqBody string `json:"reqBody,omitempty" yaml:"reqBody,omitempty"`

	// Error 字段表示请求中的任何错误
	// The Error field represents any errors in the request
	Error error `json:"error,omitempty" yaml:"error,omitempty"`

	// ErrorStack 字段表示错误的堆栈跟踪
	// The ErrorStack field represents the stack trace of the error
	ErrorStack string `json:"errorStack,omitempty" yaml:"errorStack,omitempty"`
}

// Reset 是一个方法，用于重置 LogEvent 结构体的所有字段
// Reset is a method used to reset all fields of the LogEvent struct
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
