package orbit

// Options表示应用程序的配置选项。
// Options represents the configuration options for the application.
type Options struct {
	pprof             bool // Enable pprof endpoint
	swagger           bool // Enable swagger documentation
	metric            bool // Enable metric collection
	trailingSlash     bool // Enable trailing slash redirection
	fixedPath         bool // Enable fixed path redirection
	forwordByClientIp bool // Enable client IP forwarding
	recReqBody        bool // Enable recording request body
}

// NewOptions 创建Options的新实例。
// NewOptions creates a new instance of Options.
func NewOptions() *Options {
	return &Options{}
}

// EnablePProf 启用pprof调试服务。
// EnablePProf enables the pprof debug service.
func (o *Options) EnablePProf() *Options {
	o.pprof = true
	return o
}

// EnableSwagger 启用swagger文档。
// EnableSwagger enables the swagger documentation.
func (o *Options) EnableSwagger() *Options {
	o.swagger = true
	return o
}

// EnableMetric 启用指标收集。
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

// EnableForwardedByClientIp 启用客户端IP转发。
// EnableForwardedByClientIp enables the client IP forwarding.
func (o *Options) EnableForwardedByClientIp() *Options {
	o.forwordByClientIp = true
	return o
}

// EnableRecordRequestBody 启用记录请求体。
// EnableRecordRequestBody enables the recording request body.
func (o *Options) EnableRecordRequestBody() *Options {
	o.recReqBody = true
	return o
}

// DebugOptions 返回调试选项。
// DebugOptions returns the debug options.
func DebugOptions() *Options {
	return NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()
}

// ReleaseOptions 返回发布选项。
// ReleaseOptions returns the release options.
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
