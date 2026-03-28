package httptool

import (
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

func TestGetLoggerFromContext(t *testing.T) {
	defaultLogger := &com.DefaultLogrLogger

	// Test with nil context
	result := GetLoggerFromContext(nil)
	assert.Equal(t, defaultLogger, result)

	// Test when RequestLoggerKey does not exist in the context
	context := &gin.Context{}
	result = GetLoggerFromContext(context)
	assert.Equal(t, defaultLogger, result)

	// Test when RequestLoggerKey exists with *logr.Logger
	context = &gin.Context{}
	logrLogger := com.DefaultLogrLogger
	context.Set(com.RequestLoggerKey, &logrLogger)
	result = GetLoggerFromContext(context)
	assert.NotNil(t, result)

	// Test when RequestLoggerKey exists with unsupported type
	context = &gin.Context{}
	context.Set(com.RequestLoggerKey, "unsupported")
	result = GetLoggerFromContext(context)
	assert.Equal(t, defaultLogger, result)
}
