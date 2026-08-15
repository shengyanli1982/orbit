package middleware

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestCors(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Add the Cors middleware to the router
	router.Use(Cors())

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)

	header := recorder.Header()
	assert.Empty(t, header.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, header.Get("Access-Control-Allow-Methods"))
	assert.Empty(t, header.Get("Access-Control-Allow-Headers"))
	assert.Empty(t, header.Get("Access-Control-Expose-Headers"))
	assert.Empty(t, header.Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, header.Get("Access-Control-Max-Age"))
	assert.Equal(t, "{\"message\":\"OK\"}", recorder.Body.String())
}

func TestCorsWithPolicyDisabled(t *testing.T) {
	router := gin.New()
	router.Use(CorsWithPolicy(com.CORSPolicy{Enabled: false}))
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
}

func TestCorsWithPolicyWhitelist(t *testing.T) {
	router := gin.New()
	router.Use(CorsWithPolicy(com.CORSPolicy{
		Enabled:          true,
		AllowedOrigins:   []string{"https://app.example.com"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAgeSeconds:    600,
	}))
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "https://app.example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", recorder.Header().Get("Vary"))
}

func TestCorsWithPolicyWhitelistReject(t *testing.T) {
	router := gin.New()
	router.Use(CorsWithPolicy(com.CORSPolicy{
		Enabled:        true,
		AllowedOrigins: []string{"https://app.example.com"},
	}))
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
}

func TestCors_PreflightAndEdgeCases(t *testing.T) {
	t.Run("OPTIONS_AllowAll", func(t *testing.T) {
		router := gin.New()
		router.Use(Cors())
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Empty(t, recorder.Body.String(), "OPTIONS should have empty body")
	})

	t.Run("OPTIONS_Whitelist_Match", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:          true,
			AllowedOrigins:   []string{"https://app.example.com"},
			AllowedMethods:   []string{"GET"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
			MaxAgeSeconds:    300,
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Equal(t, "https://app.example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", recorder.Header().Get("Vary"))
		assert.Equal(t, "GET", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "300", recorder.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("OPTIONS_Whitelist_Reject", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:          true,
			AllowedOrigins:   []string{"https://app.example.com"},
			AllowCredentials: true,
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS_Disabled", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{Enabled: false}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("AllowAllOrigins_WithOrigin", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:          true,
			AllowAllOrigins:  true,
			AllowedMethods:   []string{"GET"},
			AllowCredentials: false,
			MaxAgeSeconds:    0,
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "false", recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Empty(t, recorder.Header().Get("Vary"), "AllowAllOrigins should not set Vary")
		assert.Empty(t, recorder.Header().Get("Access-Control-Max-Age"), "MaxAge 0 should not emit header")
	})

	t.Run("AllowAllOrigins_NoOrigin", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:         true,
			AllowAllOrigins: true,
			AllowedMethods:  []string{"GET"},
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "{\"message\":\"OK\"}", recorder.Body.String())
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Expose-Headers"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Max-Age"))
		assert.Empty(t, recorder.Header().Get("Vary"))
	})

	t.Run("AllowAllOrigins_NoOrigin_OPTIONS", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:         true,
			AllowAllOrigins: true,
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Empty(t, recorder.Body.String())
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("FastPath_NoOrigin_NotAllowAll", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:        true,
			AllowedOrigins: []string{"https://app.example.com"},
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("FastPath_NoOrigin_NotAllowAll_OPTIONS", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:        true,
			AllowedOrigins: []string{"https://app.example.com"},
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Wildcard_AllowedOrigins", func(t *testing.T) {
		router := gin.New()
		router.Use(CorsWithPolicy(com.CORSPolicy{
			Enabled:          true,
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET"},
			AllowCredentials: true,
		}))
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://any.example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "https://any.example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", recorder.Header().Get("Vary"))
	})

	t.Run("ConcurrentRequests", func(t *testing.T) {
		router := gin.New()
		router.Use(Cors())
		router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "OK"}) })

		const goroutines = 32
		const iterations = 100
		var wg sync.WaitGroup
		wg.Add(goroutines)
		errCh := make(chan string, goroutines*iterations)
		for range goroutines {
			go func() {
				defer wg.Done()
				for i := range iterations {
					method := http.MethodGet
					if i%5 == 0 {
						method = http.MethodOptions
					}
					req, _ := http.NewRequest(method, "/test", nil)
					req.Header.Set("Origin", "https://app.example.com")
					recorder := httptest.NewRecorder()
					router.ServeHTTP(recorder, req)
					if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
						errCh <- fmt.Sprintf("unexpected Allow-Origin: %q", recorder.Header().Get("Access-Control-Allow-Origin"))
						return
					}
				}
			}()
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			t.Fatal(err)
		}
	})
}

func TestFormatDurationMs(t *testing.T) {
	baseline := func(ns int64) string {
		ms := float64(ns) / 1e6
		return strconv.FormatFloat(math.Round(ms*100)/100, 'f', -1, 64) + "ms"
	}

	cases := []int64{
		0,
		1,
		500_000,
		999_999,
		1_000_000,
		1_234_567,
		1_500_000,
		123_456_789,
		1_234_567_890_123,
		math.MaxInt64 / 4,
	}
	for _, ns := range cases {
		assert.Equal(t, baseline(ns), formatDurationMs(ns), "ns=%d", ns)
	}
}

func TestAccessLogger(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewZapLogger(zapcore.AddSync(buff), false).GetLogrLogger()

	// Create a test log event function
	logEventFunc := func(logger *logr.Logger, event *log.LogEvent) {
		logger.Info(event.Message)
	}

	// Add the AccessLogger middleware to the router
	router.Use(AccessLogger(logger, logEventFunc, true))

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "{\"message\":\"OK\"}", recorder.Body.String())

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server access log", "buffer should contain the message")
}

func TestLogrAccessLogger(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewLogrLogger(buff, false).GetLogrLogger()

	// Create a test log event function
	logEventFunc := func(logger *logr.Logger, event *log.LogEvent) {
		logger.Info(event.Message, "code", event.Code, "method", event.Method, "path", event.Path)
	}

	// Add the AccessLogger middleware to the router
	router.Use(AccessLogger(logger, logEventFunc, true))

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 200
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "{\"message\":\"OK\"}", recorder.Body.String())

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server access log", "buffer should contain the message")
}

func TestAccessLoggerSkipsInternalResources(t *testing.T) {
	router := gin.New()

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	}

	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewZapLogger(zapcore.AddSync(buff), false).GetLogrLogger()

	loggedPaths := make([]string, 0, 2)
	logEventFunc := func(logger *logr.Logger, event *log.LogEvent) {
		loggedPaths = append(loggedPaths, event.Path)
		logger.Info(event.Message, "path", event.Path)
	}

	router.Use(AccessLogger(logger, logEventFunc, true))
	router.GET(com.HealthCheckURLPath, handler)
	router.GET("/test", handler)

	req, _ := http.NewRequest(http.MethodGet, com.HealthCheckURLPath, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, loggedPaths, "internal resource path must not emit access log event")

	req, _ = http.NewRequest(http.MethodGet, "/test", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, []string{"/test"}, loggedPaths)
	assert.Contains(t, buff.String(), "http server access log", "buffer should contain the message")
}

func TestRecovery(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		panic("test panic")
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewZapLogger(zapcore.AddSync(buff), false).GetLogrLogger()

	// Add the Recovery middleware to the router
	router.Use(Recovery(logger, log.DefaultRecoveryEventFunc))

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 500
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "[500] http server internal error, method: GET, path: /test", recorder.Body.String())

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server recovery from panic", "buffer should contain the message")
}

func TestLogrRecovery(t *testing.T) {
	// Create a new Gin router
	router := gin.New()

	// Create a test handler
	handler := func(c *gin.Context) {
		panic("test panic")
	}

	// Create a test logger
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := log.NewLogrLogger(buff, false).GetLogrLogger()

	// Add the Recovery middleware to the router
	router.Use(Recovery(logger, log.DefaultRecoveryEventFunc))

	// Add the test handler to the router
	router.GET("/test", handler)

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Assert that the response status code is 500
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "[500] http server internal error, method: GET, path: /test", recorder.Body.String())

	// Print the log buffer
	fmt.Println(buff.String())

	// Assert that the log buffer contains the expected message
	assert.Contains(t, buff.String(), "http server recovery from panic", "buffer should contain the message")
}

func TestHeaderFirstValue(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		key    string
		want   string
	}{
		{
			name:   "present single value",
			header: http.Header{com.HttpHeaderContentType: []string{"application/json"}},
			key:    com.HttpHeaderContentType,
			want:   "application/json",
		},
		{
			name:   "missing key",
			header: http.Header{},
			key:    com.HttpHeaderRequestID,
			want:   "",
		},
		{
			name:   "multiple values takes first",
			header: http.Header{com.HttpHeaderForwardedFor: []string{"10.0.0.1", "10.0.0.2"}},
			key:    com.HttpHeaderForwardedFor,
			want:   "10.0.0.1",
		},
		{
			name:   "empty slice",
			header: http.Header{com.HttpHeaderRequestID: {}},
			key:    com.HttpHeaderRequestID,
			want:   "",
		},
		{
			name:   "nil header",
			header: nil,
			key:    "Origin",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := headerFirstValue(tt.header, tt.key)
			assert.Equal(t, tt.want, got)
			if tt.header != nil {
				assert.Equal(t, tt.header.Get(tt.key), got)
			}
		})
	}
}

func TestAccessLoggerMultiValueHeaders(t *testing.T) {
	var gotID, gotAgent, gotForwarded, gotContentType string

	logEventFunc := func(_ *logr.Logger, event *log.LogEvent) {
		gotID = event.ID
		gotAgent = event.Agent
		gotForwarded = event.ForwardedFor
		gotContentType = event.ReqContentType
	}

	logger := logr.Discard()

	router := gin.New()
	router.Use(AccessLogger(&logger, logEventFunc, false))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header[com.HttpHeaderForwardedFor] = []string{"10.0.0.1", "10.0.0.2"}
	req.Header[com.HttpHeaderRequestID] = []string{"req-first", "req-second"}
	req.Header["User-Agent"] = []string{"orbit-agent", "orbit-agent-2"}
	req.Header[com.HttpHeaderContentType] = []string{"application/json", "text/plain"}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "10.0.0.1", gotForwarded)
	assert.Equal(t, req.Header.Get(com.HttpHeaderForwardedFor), gotForwarded)
	assert.Equal(t, "req-first", gotID)
	assert.Equal(t, "orbit-agent", gotAgent)
	assert.Equal(t, "application/json", gotContentType)
}
