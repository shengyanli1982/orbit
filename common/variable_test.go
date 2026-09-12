package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSugaredLoggerAlias(t *testing.T) {
	// 新旧名称必须指向同一日志记录器实例（别名同值契约）
	assert.Same(t, DefaultSugeredLogger, DefaultSugaredLogger)
}
