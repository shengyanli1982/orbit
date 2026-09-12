package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
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

	router.Use(Recovery(logger, logEventFunc, false))
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

// TestRecovery_DoesNotRecordBodyWhenRecordingDisabled 验证 #4：未启用请求体记录时，
// Recovery 的 panic 日志事件不得包含请求体——旧实现无条件读取请求体，绕过了
// recReqBody 开关与 CanRecordContextBody 的内容类型过滤（隐私/合规缺陷）。
func TestRecovery_DoesNotRecordBodyWhenRecordingDisabled(t *testing.T) {
	router := gin.New()
	logger := logr.Discard()

	var gotReqBody string
	logEventFunc := func(_ *logr.Logger, event *log.LogEvent) {
		gotReqBody = event.ReqBody
	}

	router.Use(Recovery(&logger, logEventFunc, false))
	router.POST("/panic", func(c *gin.Context) { panic("boom") })

	req, _ := http.NewRequest(http.MethodPost, "/panic", strings.NewReader(`{"token":"super-secret"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Empty(t, gotReqBody, "recovery event must not contain the request body when body recording is disabled")
}

// TestRecovery_RecordsBodyWhenEnabled 验证启用记录且内容类型通过 CanRecordContextBody
// 过滤时，panic 事件仍记录请求体（正例对照，防止门控过度）。
func TestRecovery_RecordsBodyWhenEnabled(t *testing.T) {
	router := gin.New()
	logger := logr.Discard()

	var gotReqBody string
	logEventFunc := func(_ *logr.Logger, event *log.LogEvent) {
		gotReqBody = event.ReqBody
	}

	router.Use(Recovery(&logger, logEventFunc, true))
	router.POST("/panic", func(c *gin.Context) { panic("boom") })

	req, _ := http.NewRequest(http.MethodPost, "/panic", strings.NewReader(`{"data":"payload-data"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, gotReqBody, "payload-data", "recovery event must record the request body when recording is enabled and content type is recordable")
}

// TestRecovery_SkipsNonRecordableContentType 验证启用记录但内容类型未通过
// CanRecordContextBody 过滤（如二进制流）时，panic 事件不读取请求体。
func TestRecovery_SkipsNonRecordableContentType(t *testing.T) {
	router := gin.New()
	logger := logr.Discard()

	var gotReqBody string
	logEventFunc := func(_ *logr.Logger, event *log.LogEvent) {
		gotReqBody = event.ReqBody
	}

	router.Use(Recovery(&logger, logEventFunc, true))
	router.POST("/panic", func(c *gin.Context) { panic("boom") })

	req, _ := http.NewRequest(http.MethodPost, "/panic", strings.NewReader("binary-blob-secret"))
	req.Header.Set("Content-Type", "application/octet-stream")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Empty(t, gotReqBody, "recovery event must skip bodies whose content type fails CanRecordContextBody")
}
