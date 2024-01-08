package orbit

type Options struct {
	pprof             bool
	swagger           bool
	trailingSlash     bool
	fixedPath         bool
	forwordByClientIp bool
}

func NewOptions() *Options {
	return &Options{
		pprof:         true,
		swagger:       false,
		trailingSlash: false,
		fixedPath:     false,
	}
}

func (o *Options) EnablePProf() *Options {
	o.pprof = true
	return o
}

func (o *Options) EnableSwagger() *Options {
	o.swagger = true
	return o
}

func (o *Options) EnableRedirectTrailingSlash() *Options {
	o.trailingSlash = true
	return o
}

func (o *Options) EnableRedirectFixedPath() *Options {
	o.fixedPath = true
	return o
}

func (o *Options) EnableForwardedByClientIp() *Options {
	o.forwordByClientIp = true
	return o
}
