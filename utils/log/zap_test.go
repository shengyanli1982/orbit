package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestGetZapLogger(t *testing.T) {
	// Create a new buffer
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	// Create a new logger
	logger := NewZapLogger(zapcore.AddSync(buff), false)
	zapLogger := logger.GetZapLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, zapLogger, "zapLogger should not be nil")
	assert.Equal(t, zapLogger.Core().Enabled(zap.DebugLevel), true, "zapLogger should be at Debug level")

	// Log a message
	zapLogger.Debug("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}

func TestGetZapSugaredLogger(t *testing.T) {
	// Create a new buffer
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	// Create a new logger
	logger := NewZapLogger(zapcore.AddSync(buff), false)
	sugaredLogger := logger.GetZapSugaredLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, sugaredLogger, "sugaredLogger should not be nil")

	// Log a message
	sugaredLogger.Debug("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}

// TestNewZapLoggerDefaultIncludesCaller 钉住默认行为基线：
// NewZapLogger 不传任何选项时，日志输出必须包含 caller 字段（与历史行为一致）
func TestNewZapLoggerDefaultIncludesCaller(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	logger := NewZapLogger(zapcore.AddSync(buff), false)
	logger.GetZapLogger().Info("test message")

	assert.Contains(t, buff.String(), `"caller":"log/zap_test.go:`, "default logger output should contain the caller field")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}

// TestNewZapLoggerWithLogCallerDisabled 验证行为开关：
// 传入 WithLogCaller(false) 后日志输出不再包含 caller 字段（省去每条日志的栈捕获），
// 其余字段与默认行为一致
func TestNewZapLoggerWithLogCallerDisabled(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	logger := NewZapLogger(zapcore.AddSync(buff), false, WithLogCaller(false))
	logger.GetZapLogger().Info("test message")

	assert.NotContains(t, buff.String(), `"caller":`, "logger built with WithLogCaller(false) should omit the caller field")
	assert.Contains(t, buff.String(), `"level":"INFO"`, "other fields should remain unchanged")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}

func TestGetZapStdLogger(t *testing.T) {
	// Create a new buffer
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	// Create a new logger
	logger := NewZapLogger(zapcore.AddSync(buff), false)
	stdLogger := logger.GetStandardLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, stdLogger, "stdLogger should not be nil")

	// Log a message
	stdLogger.Print("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}
