package httptool

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

// errReader 先返回部分数据再返回错误，用于模拟 io.Copy 读取请求体失败的场景
type errReader struct {
	data []byte
	err  error
}

func (r *errReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	return 0, r.err
}

func TestGenerateRequestBody(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte("test body")
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	context.Request = request

	// Repeat read the request body 100 times
	for range 100 {
		// Call the GenerateRequestBody function
		body, err := GenerateRequestBody(context)

		// Assert that there is no error
		assert.NoError(t, err)

		// Assert that the returned body matches the original request body
		assert.Equal(t, requestBody, body)
	}

	// Assert that the request body has been replaced with the buffer
	bufferedBody, _ := io.ReadAll(context.Request.Body)
	assert.Equal(t, requestBody, bufferedBody)
}

func TestParseRequestBodyJSON(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte(`{"test": "body"}`)
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", binding.MIMEJSON)
	context.Request = request

	// Call the ParseRequestBody function
	var value interface{}
	err := ParseRequestBody(context, &value, false)

	// Assert that there is no error
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Assert that the returned body matches the original request body
	assert.Equal(t, map[string]interface{}{"test": "body"}, value)
}

func TestParseRequestBodyYAML(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte("test: body")
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", binding.MIMEYAML)
	context.Request = request

	// Call the ParseRequestBody function
	var value interface{}
	err := ParseRequestBody(context, &value, false)

	// Assert that there is no error
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Assert that the returned body matches the original request body
	assert.Equal(t, map[string]interface{}{"test": "body"}, value)
}

type testXmlStruct struct {
	XMLName xml.Name `xml:"block"`
	Test    string   `xml:"test"`
}

func TestParseRequestBodyXML(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte(`<block><test>body</test></block>`)
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", binding.MIMEXML)
	context.Request = request

	// Call the ParseRequestBody function
	var value testXmlStruct
	err := ParseRequestBody(context, &value, false)

	// Assert that there is no error
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Assert that the returned body matches the original request body
	assert.Equal(t, testXmlStruct{XMLName: xml.Name{Space: "", Local: "block"}, Test: "body"}, value)
}

type testFormStruct struct {
	Test string `form:"test"`
}

func TestParseRequestBodyForm(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte("test=body")
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", binding.MIMEPOSTForm)
	context.Request = request

	// Call the ParseRequestBody function
	var value testFormStruct
	err := ParseRequestBody(context, &value, false)

	// Assert that there is no error
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Assert that the returned body matches the original request body
	assert.Equal(t, testFormStruct{Test: "body"}, value)
}

func TestParseRequestBodyProtoBuf(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte{0x0A, 0x04, 0x74, 0x65, 0x73, 0x74}
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", binding.MIMEPROTOBUF)
	context.Request = request

	// Call the ParseRequestBody function
	var value TestProtoBufStruct
	err := ParseRequestBody(context, &value, false)

	// Assert that there is no error
	assert.NoError(t, err)
	assert.NotNil(t, &value)

	// Assert that the returned body matches the original request body
	assert.Equal(t, "test", value.Test)
}

func TestParseRequestBodyInvalidContentType(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte("test body")
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", "invalid")
	context.Request = request

	// Call the ParseRequestBody function
	var value interface{}
	err := ParseRequestBody(context, &value, false)

	// Assert that there is no error
	assert.NoError(t, err)
	assert.Nil(t, value)
}

func TestParseRequestBodyEmptyBody(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with an empty body
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	context.Request = request

	// Call the ParseRequestBody function with emptyRequestBodyContent set to true
	var value interface{}
	err := ParseRequestBody(context, &value, true)

	// Assert that there is no error
	assert.Equal(t, err, ErrorContentTypeIsEmpty)

	// Assert that the request body has been replaced with the buffer
	bufferedBody, _ := io.ReadAll(context.Request.Body)
	assert.Equal(t, []byte{}, bufferedBody)
}

func TestGenerateRequestPath(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with an empty body
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	context.Request = request

	// Call the GenerateRequestPath function
	path := GenerateRequestPath(context)

	// Assert that the returned path matches the expected path
	assert.Equal(t, "/test", path)

	// Set the request URL path with a query string
	context.Request.URL.RawQuery = "param=value"

	// Call the GenerateRequestPath function
	path = GenerateRequestPath(context)

	// Assert that the returned path matches the expected path with the query string
	assert.Equal(t, "/test?param=value", path)
}

func TestStringFilterFlags(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Test case 1",
			input:    "abc; def",
			expected: "abc",
		},
		{
			name:     "Test case 2",
			input:    "xyz",
			expected: "xyz",
		},
		{
			name:     "Test case 3",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := StringFilterFlags(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalcRequestSize(t *testing.T) {
	// Create a new request
	request, _ := http.NewRequest(http.MethodGet, "/ping", nil)

	// Calculate the request size
	size := CalcRequestSize(request)

	// Assert that the size is correct
	assert.Equal(t, int64(16), size)
}

func TestGenerateRequestBodyCopyFailureDoesNotLeakPartialBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	// 注入读取失败的 Body：io.Copy 先拷入 "partial" 再报错
	request.Body = io.NopCloser(&errReader{data: []byte("partial"), err: errors.New("read failure")})
	context.Request = request

	// 第一次调用：io.Copy 失败，返回错误
	body, err := GenerateRequestBody(context)
	assert.Error(t, err)
	assert.Equal(t, []byte("failed to get request body"), body)

	// 第二次调用：换上正常请求体，不得把上次失败的残缺数据当作完整请求体返回
	request.Body = io.NopCloser(strings.NewReader("complete"))
	body, err = GenerateRequestBody(context)
	assert.NoError(t, err)
	assert.Equal(t, []byte("complete"), body)
}

func TestGenerateRequestBodyErrorPathReturnsWritableBytes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	request.Body = nil
	context.Request = request

	body, err := GenerateRequestBody(context)
	assert.NoError(t, err)
	assert.Equal(t, []byte("request body is nil"), body)

	// 返回的切片必须可写：若别名字符串常量的只读数据段，此处写入会 segfault
	body[0] = 'x'
	assert.Equal(t, byte('x'), body[0])
}
