package log

// LogEvent represents a log event.
type LogEvent struct {
	Message        string `json:"message,omitempty" yaml:"message,omitempty"`               // Message contains the log message.
	ID             string `json:"id,omitempty" yaml:"id,omitempty"`                         // ID contains the unique identifier of the log event.
	IP             string `json:"ip,omitempty" yaml:"ip,omitempty"`                         // IP contains the IP address of the client.
	EndPoint       string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`             // EndPoint contains the endpoint of the request.
	Path           string `json:"path,omitempty" yaml:"path,omitempty"`                     // Path contains the path of the request.
	Method         string `json:"method,omitempty" yaml:"method,omitempty"`                 // Method contains the HTTP method of the request.
	Code           int    `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`         // Code contains the HTTP status code of the response.
	Status         string `json:"status,omitempty" yaml:"status,omitempty"`                 // Status contains the status message of the response.
	Latency        string `json:"latency,omitempty" yaml:"latency,omitempty"`               // Latency contains the request latency.
	Agent          string `json:"agent,omitempty" yaml:"agent,omitempty"`                   // Agent contains the user agent of the client.
	ReqContentType string `json:"reqContentType,omitempty" yaml:"reqContentType,omitempty"` // ReqContentType contains the content type of the request.
	ReqQuery       string `json:"query,omitempty" yaml:"query,omitempty"`                   // ReqQuery contains the query parameters of the request.
	ReqBody        string `json:"reqBody,omitempty" yaml:"reqBody,omitempty"`               // ReqBody contains the request body.
	Error          any    `json:"error,omitempty" yaml:"error,omitempty"`                   // Error contains the error object.
	ErrorStack     string `json:"errorStack,omitempty" yaml:"errorStack,omitempty"`         // ErrorStack contains the stack trace of the error.
}

// Reset resets the LogEvent fields to their zero values.
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
