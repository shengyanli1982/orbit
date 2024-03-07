package orbit

// Options 表示应用程序的配置选项。
// Options represents the configuration options for the application.
type Options struct {
	pprof             bool // 启用 pprof 端点
	swagger           bool // 启用 swagger 文档
	metric            bool // 启用度量收集
	trailingSlash     bool // 启用尾部斜杠重定向
	fixedPath         bool // 启用固定路径重定向
	forwordByClientIp bool // 启用客户端 IP 转发
	recReqBody        bool // 启用请求体记录
}

// NewOptions 创建一个新的 Options 实例。
// NewOptions creates a new instance of Options.
func NewOptions() *Options {
	return &Options{}
}

// EnablePProf 启用 pprof 端点。
// EnablePProf enables the pprof endpoint.
func (o *Options) EnablePProf() *Options {
	o.pprof = true
	return o
}

// EnableSwagger 启用 swagger 文档。
// EnableSwagger enables the swagger documentation.
func (o *Options) EnableSwagger() *Options {
	o.swagger = true
	return o
}

// EnableMetric 启用度量收集。
// EnableMetric enables the metric collection.
func (o *Options) EnableMetric() *Options {
	o.metric = true
	return o
}

// EnableRedirectTrailingSlash 启用尾部斜杠重定向。
// EnableRedirectTrailingSlash enables the trailing slash redirection.
func (o *Options) EnableRedirectTrailingSlash() *Options {
	o.trailingSlash = true
	return o
}

// EnableRedirectFixedPath 启用固定路径重定向。
// EnableRedirectFixedPath enables the fixed path redirection.
func (o *Options) EnableRedirectFixedPath() *Options {
	o.fixedPath = true
	return o
}

// EnableForwardedByClientIp 启用客户端 IP 转发。
// EnableForwardedByClientIp enables the client IP forwarding.
func (o *Options) EnableForwardedByClientIp() *Options {
	o.forwordByClientIp = true
	return o
}

// EnableRecordRequestBody 启用请求体记录。
// EnableRecordRequestBody enables the recording request body.
func (o *Options) EnableRecordRequestBody() *Options {
	o.recReqBody = true
	return o
}

// DebugOptions 返回一个 Options 实例，该实例启用了 pprof、swagger、metric 和请求体记录功能，用于调试环境。
// DebugOptions returns an Options instance that enables pprof, swagger, metric, and request body recording, for debugging environment.
func DebugOptions() *Options {
	return NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()
}

// ReleaseOptions 返回一个 Options 实例，该实例仅启用了 metric 功能，用于生产环境。
// ReleaseOptions returns an Options instance that only enables the metric function, for production environment.
func ReleaseOptions() *Options {
	return NewOptions().EnableMetric()
}

// isOptionsValid 检查选项是否有效，并在必要时应用默认值。
// isOptionsValid checks if the options is valid and applies default values if necessary.
func isOptionsValid(opts *Options) *Options {
	if opts == nil {
		opts = DebugOptions()
	}
	return opts
}
