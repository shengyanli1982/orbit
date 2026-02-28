package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit/utils/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestRecovery_EventCodeIs500(t *testing.T) {
	router := gin.New()

	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewZapLogger(zapcore.AddSync(buff), false).GetLogrLogger()

	var gotCode int
	var gotStatus string
	logEventFunc := func(_ *logr.Logger, event *log.LogEvent) {
		gotCode = event.Code
		gotStatus = event.Status
	}

	router.Use(Recovery(logger, logEventFunc))
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, http.StatusInternalServerError, gotCode)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), gotStatus)
}
