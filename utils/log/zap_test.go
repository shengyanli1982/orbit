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
	logger := NewZapLogger(zapcore.AddSync(buff))
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
	logger := NewZapLogger(zapcore.AddSync(buff))
	sugaredLogger := logger.GetZapSugaredLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, sugaredLogger, "sugaredLogger should not be nil")

	// Log a message
	sugaredLogger.Debug("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}

func TestGetZapStdLogger(t *testing.T) {
	// Create a new buffer
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	// Create a new logger
	logger := NewZapLogger(zapcore.AddSync(buff))
	stdLogger := logger.GetStdLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, stdLogger, "stdLogger should not be nil")

	// Log a message
	stdLogger.Print("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}
