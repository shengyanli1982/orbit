package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

func TestBodyBuffer(t *testing.T) {
	// 设置 gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 生成大量数据用于测试
	largeData := make([]byte, 1024*1024) // 1MB 数据
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	tests := []struct {
		name           string
		requestBody    []byte
		responseBody   string
		expectedStatus int
	}{
		{
			name:           "Test normal request with body",
			requestBody:    []byte("test request body"),
			responseBody:   "test response body",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test empty request",
			requestBody:    []byte(""),
			responseBody:   "empty request test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test large data request",
			requestBody:    largeData,
			responseBody:   "large data processed",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test special characters",
			requestBody:    []byte("特殊字符测试：!@#$%^&*()_+\n\t中文测试"),
			responseBody:   "special chars processed",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test error response",
			requestBody:    []byte("error trigger"),
			responseBody:   "internal server error",
			expectedStatus: http.StatusInternalServerError,
		},
		// {
		// 	name:           "Test binary data",
		// 	requestBody:    []byte{0x00, 0x01, 0x02, 0x03},
		// 	responseBody:   "binary data processed",
		// 	expectedStatus: http.StatusOK,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个新的 gin 引擎
			router := gin.New()

			// 添加 BodyBuffer 中间件
			router.Use(BodyBuffer())

			// 设置测试路由
			router.POST("/test", func(c *gin.Context) {
				// 验证缓冲区是否正确设置
				reqBuffer, exists := c.Get(com.RequestBodyBufferKey)
				assert.True(t, exists)
				assert.NotNil(t, reqBuffer)

				// 验证响应体缓冲区是否正确设置
				respBuffer, exists := c.Get(com.ResponseBodyBufferKey)
				assert.True(t, exists)
				assert.NotNil(t, respBuffer)

				// 读取请求体
				body, err := io.ReadAll(c.Request.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody, body)

				// 对于大数据测试，添加长度验证
				if tt.name == "Test large data request" {
					assert.Equal(t, len(largeData), len(body))
				}

				// 测试多次读取请求体
				if tt.name == "Test binary data" {
					// 第二次读取应该仍然可以得到完整数据
					body2, err := io.ReadAll(c.Request.Body)
					assert.NoError(t, err)
					assert.Equal(t, tt.requestBody, body2)
				}

				// 写入响应
				c.String(tt.expectedStatus, tt.responseBody)
			})

			// 使用字节切片创建请求
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(tt.requestBody))
			resp := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(resp, req)

			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// 验证响应体
			assert.Equal(t, tt.responseBody, resp.Body.String())
		})
	}
}
