package orbit

import (
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// WithSugaredLogger 为 Config 实例设置一个新的 sugared logger
//
// Deprecated: Use WithLogger instead (since v0.2.9). This method will be removed in the next release.
func (c *Config) WithSugaredLogger(logger *zap.SugaredLogger) *Config {
	l := zapr.NewLogger(logger.Desugar())
	c.logger = &l
	return c
}
