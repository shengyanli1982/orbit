package orbit

// Options 表示应用程序的配置选项
type Options struct {
	healthCheck       bool // 启用健康检查
	pprof             bool // 启用 pprof 端点
	swagger           bool // 启用 swagger 文档
	metric            bool // 启用度量收集
	trailingSlash     bool // 启用尾部斜杠重定向
	fixedPath         bool // 启用固定路径重定向
	forwordByClientIp bool // 启用客户端 IP 转发
	recReqBody        bool // 启用请求体记录
}

// NewOptions 创建一个新的 Options 实例
func NewOptions() *Options {
	return &Options{healthCheck: true}
}

// EnableHealthCheck 启用健康检查
func (o *Options) EnableHealthCheck() *Options {
	o.healthCheck = true
	return o
}

// EnablePProf 启用 pprof 端点
func (o *Options) EnablePProf() *Options {
	o.pprof = true
	return o
}

// EnableSwagger 启用 swagger 文档
func (o *Options) EnableSwagger() *Options {
	o.swagger = true
	return o
}

// EnableMetric 启用度量收集
func (o *Options) EnableMetric() *Options {
	o.metric = true
	return o
}

// EnableRedirectTrailingSlash 启用尾部斜杠重定向
func (o *Options) EnableRedirectTrailingSlash() *Options {
	o.trailingSlash = true
	return o
}

// EnableRedirectFixedPath 启用固定路径重定向
func (o *Options) EnableRedirectFixedPath() *Options {
	o.fixedPath = true
	return o
}

// EnableForwardedByClientIp 启用客户端 IP 转发
func (o *Options) EnableForwardedByClientIp() *Options {
	o.forwordByClientIp = true
	return o
}

// EnableRecordRequestBody 启用请求体记录
func (o *Options) EnableRecordRequestBody() *Options {
	o.recReqBody = true
	return o
}

// DebugOptions 返回一个启用了 pprof、swagger、metric 和请求体记录功能的 Options 实例，用于调试环境
func DebugOptions() *Options {
	return NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()
}

// ReleaseOptions 返回一个仅启用了 metric 功能的 Options 实例，用于生产环境
func ReleaseOptions() *Options {
	return NewOptions().EnableMetric()
}

// EmptyOptions 返回一个空的 Options 实例
func EmptyOptions() *Options {
	return &Options{}
}

// isOptionsValid 检查选项是否有效，并在必要时应用默认值
func isOptionsValid(opts *Options) *Options {
	if opts == nil {
		opts = DebugOptions()
	}
	return opts
}
