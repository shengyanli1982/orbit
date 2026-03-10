package log

import (
	"testing"

	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestDefaultAccessEventFuncUsesStandardizedKeys(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zapr.NewLogger(zap.New(core))

	event := &LogEvent{
		Message:      "http server access log",
		ID:           "req-1",
		IP:           "1.2.3.4",
		ForwardedFor: "5.6.7.8, 10.0.0.1",
		EndPoint:     "10.0.0.10:12345",
		Path:         "/demo",
		Method:       "GET",
		Code:         200,
		Status:       "OK",
		Latency:      "25ms",
		LatencyMs:    25,
		Agent:        "curl/8.0",
	}

	DefaultAccessEventFunc(&logger, event)

	entries := observed.AllUntimed()
	if assert.Len(t, entries, 1) {
		fields := entries[0].ContextMap()
		assert.Equal(t, "req-1", fields["request_id"])
		assert.Equal(t, "1.2.3.4", fields["client_ip"])
		assert.Equal(t, "5.6.7.8, 10.0.0.1", fields["forwarded_for"])
		assert.Equal(t, int64(25), fields["latency_ms"])
	}
}
