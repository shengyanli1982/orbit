package httptool

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestGenerateRequestBody(t *testing.T) {
	// Create a new Gin context
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with a sample body
	requestBody := []byte("test body")
	request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	context.Request = request

	// Repeat read the request body 100 times
	for i := 0; i < 100; i++ {
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

func TestProtoBufXXX(t *testing.T) {
	test := TestProtoBufStruct{Test: "test"}

	protoBuf, _ := proto.Marshal(&test)

	fmt.Println(protoBuf)
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
