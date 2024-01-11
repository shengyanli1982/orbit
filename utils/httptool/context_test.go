package httptool

import (
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetLoggerFromContext(t *testing.T) {
	// Create a mock gin.Context
	context := &gin.Context{}

	// Test when RequestLoggerKey exists in the context
	logger := &zap.SugaredLogger{}
	context.Set(com.RequestLoggerKey, logger)
	result := GetLoggerFromContext(context)
	assert.Equal(t, logger, result)

	// Test when RequestLoggerKey does not exist in the context
	result = GetLoggerFromContext(context)
	assert.Equal(t, logger, result)
}
