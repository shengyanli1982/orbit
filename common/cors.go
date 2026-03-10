package common

// CORSPolicy 定义跨域策略
type CORSPolicy struct {
	Enabled          bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowAllOrigins  bool     `json:"allowAllOrigins,omitempty" yaml:"allowAllOrigins,omitempty"`
	AllowedOrigins   []string `json:"allowedOrigins,omitempty" yaml:"allowedOrigins,omitempty"`
	AllowedMethods   []string `json:"allowedMethods,omitempty" yaml:"allowedMethods,omitempty"`
	AllowedHeaders   []string `json:"allowedHeaders,omitempty" yaml:"allowedHeaders,omitempty"`
	ExposeHeaders    []string `json:"exposeHeaders,omitempty" yaml:"exposeHeaders,omitempty"`
	AllowCredentials bool     `json:"allowCredentials,omitempty" yaml:"allowCredentials,omitempty"`
	MaxAgeSeconds    int      `json:"maxAgeSeconds,omitempty" yaml:"maxAgeSeconds,omitempty"`
}
