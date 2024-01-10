package orbit

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

// NewOptions creates a new instance of Options.
func NewOptions() *Options {
	return &Options{}
}

// EnablePProf enables the pprof endpoint.
func (o *Options) EnablePProf() *Options {
	o.pprof = true
	return o
}

// EnableSwagger enables the swagger documentation.
func (o *Options) EnableSwagger() *Options {
	o.swagger = true
	return o
}

// EnableMetric enables the metric collection.
func (o *Options) EnableMetric() *Options {
	o.metric = true
	return o
}

// EnableRedirectTrailingSlash enables the trailing slash redirection.
func (o *Options) EnableRedirectTrailingSlash() *Options {
	o.trailingSlash = true
	return o
}

// EnableRedirectFixedPath enables the fixed path redirection.
func (o *Options) EnableRedirectFixedPath() *Options {
	o.fixedPath = true
	return o
}

// EnableForwardedByClientIp enables the client IP forwarding.
func (o *Options) EnableForwardedByClientIp() *Options {
	o.forwordByClientIp = true
	return o
}

// EnableRecordRequestBody enables the recording request body.
func (o *Options) EnableRecordRequestBody() *Options {
	o.recReqBody = true
	return o
}

// DebugOptions returns the debug options.
func DebugOptions() *Options {
	return NewOptions().EnablePProf().EnableSwagger().EnableMetric().EnableRecordRequestBody()
}

// ReleaseOptions returns the release options.
func ReleaseOptions() *Options {
	return NewOptions().EnableMetric()
}

// isOptionsValid checks if the options is valid and applies default values if necessary.
func isOptionsValid(opts *Options) *Options {
	if opts == nil {
		opts = DebugOptions()
	}
	return opts
}
