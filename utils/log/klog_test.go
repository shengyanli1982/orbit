package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogrLogger(t *testing.T) {
	// Create a new buffer
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	// Create a new logger
	logger := NewLogrLogger(buff)
	logrLogger := logger.GetLogrLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, logrLogger, "logrLogger should not be nil")

	// Log a message
	logrLogger.Info("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}

func TestGetLogrStdLogger(t *testing.T) {
	// Create a new buffer
	buff := bytes.NewBuffer(make([]byte, 0, 1024))

	// Create a new logger
	logger := NewLogrLogger(buff)
	stdLogger := logger.GetStandardLogger()

	// Assert that the logger is not nil
	assert.NotNil(t, stdLogger, "stdLogger should not be nil")

	// Log a message
	stdLogger.Print("test message")
	assert.Contains(t, buff.String(), "test message", "buffer should contain the message")
}
