package httptool

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/internal/conver"
)

// ErrorContentTypeIsEmpty 是一个错误，表示内容类型为空。
// ErrorContentTypeIsEmpty is the error that indicates the content type is empty.
var ErrorContentTypeIsEmpty = errors.New("content type is empty")

// contentTypes 是一个字符串切片，表示 HTTP 请求支持的内容类型。
// contentTypes is a slice of strings that represents the supported content types for HTTP requests.
var contentTypes = []string{
	com.HttpHeaderJSONContentTypeValue,       // JSON 内容类型
	com.HttpHeaderJavascriptContentTypeValue, // JavaScript 内容类型
	com.HttpHeaderTextContentTypeValue,       // 文本内容类型
	com.HttpHeaderXMLContentTypeValue,        // XML 内容类型
	com.HttpHeaderPXMLContentTypeValue,       // 测试 XML 内容类型
	com.HttpHeaderYAMLContentTypeValue,       // YAML 内容类型
	com.HttpHeaderTOMLContentTypeValue,       // TOML 内容类型
}

// CalcRequestSize 函数返回请求对象的大小
// The CalcRequestSize function returns the size of the request object
func CalcRequestSize(request *http.Request) int64 {
	// 初始化 size 变量为 0
	// Initialize the size variable to 0
	size := 0

	// 如果请求的 URL 不为空，计算 URL 字符串的长度并添加到 size 变量
	// If the URL of the request is not nil, calculate the length of the URL string and add it to the size variable
	if request.URL != nil {
		size += len(request.URL.String())
	}

	// 将请求的方法和协议的长度添加到 size 变量
	// Add the length of the request's method and protocol to the size variable
	size += len(request.Method)
	size += len(request.Proto)

	// 遍历请求的头部，计算每个键值对的大小，并将其添加到 size 变量
	// Iterate through the headers of the request, calculate the size of each key-value pair, and add it to the size variable
	for name, values := range request.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}

	// 将请求的主机名（Host）的长度添加到 size 变量
	// Add the length of the request's host name (Host) to the size variable
	size += len(request.Host)

	// 如果 ContentLength 不是 -1，将 ContentLength 添加到 size 变量
	// If ContentLength is not -1, add ContentLength to the size variable
	if request.ContentLength != -1 {
		size += int(request.ContentLength)
	}

	// 返回请求的大小
	// Return the size of the request
	return int64(size)
}

// StringFilterFlags 函数返回给定字符串中的第一个标记
// The StringFilterFlags function returns the first token in the given string
func StringFilterFlags(content string) string {
	// 返回字符串中第一个 ';' 或 ' ' 之前的所有字符。如果都不存在，返回整个字符串。
	// Return all characters before the first ';' or ' ' in the string. If neither exists, return the entire string.
	if i := strings.IndexAny(content, "; "); i >= 0 {
		return content[:i]
	}
	return content
}

// CanRecordContextBody 函数检查 HTTP 请求头是否包含特定内容类型的值
// The CanRecordContextBody function checks if the HTTP request header contains a value for a specific content type
func CanRecordContextBody(header http.Header) bool {
	// 获取请求头中的内容类型
	// Get the content type from the request header
	contentType := header.Get(com.HttpHeaderContentType)

	// 如果请求头为空或内容信息不足以区分类型，直接返回 false
	// If the request header is empty or the content information is not sufficient to differentiate the type, return false directly
	if contentType == "" || !strings.Contains(contentType, "/") {
		return false
	}

	// 在 definedContentTypes 列表中查找指定的内容类型
	// Find the specified content type in the definedContentTypes list
	typeStr := StringFilterFlags(contentType)
	for _, ct := range contentTypes {
		if strings.HasPrefix(typeStr, ct) {
			return true
		}
	}

	// 如果内容类型未定义，返回 false
	// Return false if the content type is not defined
	return false
}

// GenerateRequestPath 函数从 Gin 上下文中返回请求路径
// The GenerateRequestPath function returns the request path from the Gin context
func GenerateRequestPath(context *gin.Context) string {
	// 如果请求包含查询字符串，返回整个 URL，否则返回路径
	// If the request contains a query string, return the entire URL, otherwise return the path
	if len(context.Request.URL.RawQuery) > 0 {
		return context.Request.URL.RequestURI()
	}
	return context.Request.URL.Path
}

// GenerateRequestBody 函数从 Gin 上下文中读取 HTTP 请求体，并将其存储在 Buffer Pool 对象中
// 请不要直接从 Gin 上下文中读取请求体，因为请求体只能读取一次
// The GenerateRequestBody function reads the HTTP request body from the Gin context and stores it in a Buffer Pool object
// Please don't directly read the request body from the Gin context, because the request body can only be read once
func GenerateRequestBody(context *gin.Context) ([]byte, error) {
	// 检查请求体是否为 nil
	// Check if the request body is nil
	if context.Request.Body == nil {
		return conver.StringToBytes("request body is nil"), nil
	}

	// 检查是否已经存在相关的 Buffer Pool 对象，如果不存在，创建一个新的实例
	// Check if there is already a related Buffer Pool object, if not, create a new instance
	var reqBodyBuffer *bytes.Buffer
	if buffer, ok := context.Get(com.RequestBodyBufferKey); ok {
		reqBodyBuffer = buffer.(*bytes.Buffer)
	} else {
		reqBodyBuffer = com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
	}

	// 检查 Buffer Pool 对象是否为空，如果是，读取 HTTP 请求体
	// Check if the Buffer Pool object is empty, if so, read the HTTP request body
	if reqBodyBuffer.Len() <= 0 {
		// 读取 HTTP 请求体
		// Read the HTTP request body
		body, err := io.ReadAll(context.Request.Body)
		if err != nil {
			return conver.StringToBytes("failed to get request body"), err
		}

		// 将内容写入 Buffer Pool 对象
		// Write the content to the Buffer Pool object
		_, err = reqBodyBuffer.Write(body)
		if err != nil {
			// 如果在将内容写入 Buff Pool 时发生错误，存储原始内容
			// If an error occurs while writing the content to the Buff Pool, store the original content
			context.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		} else {
			// 如果内容写入成功，将内容存储在 Buffer Pool 对象中
			// If the content is written successfully, store the content in the Buffer Pool object
			context.Request.Body = io.NopCloser(reqBodyBuffer)
		}
	}

	// 返回请求体
	// Return the request body
	return reqBodyBuffer.Bytes(), nil
}

// ParseRequestBody 函数将请求体解析到指定类型的变量 value 中，emptyRequestBodyContent 表示是否允许请求体为空
// The ParseRequestBody function parses the request body into a variable of the specified type value, emptyRequestBodyContent indicates whether an empty body is allowed
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 检查 ContentType 是否为空
	// Check if ContentType is empty
	if context.ContentType() == "" {
		return ErrorContentTypeIsEmpty
	}

	// 将请求体绑定到指定类型的 value 变量
	// Bind the request body to the specified type value
	var body []byte
	err := context.ShouldBind(value)
	if err != nil {
		// 获取请求体
		// Get the request body
		body, err = GenerateRequestBody(context)
		if err == nil {
			// 如果请求体为空且 emptyRequestBodyContent 为 true，直接返回 nil
			// If the request body is empty and emptyRequestBodyContent is true, return nil directly
			if emptyRequestBodyContent && len(body) <= 0 {
				return nil
			}
		}
	}

	// 返回错误
	// Return the error
	return err
}
