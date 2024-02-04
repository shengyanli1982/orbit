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

// ErrorContentTypeIsEmpty 是指示内容类型为空的错误。
// ErrorContentTypeIsEmpty is the error that indicates the content type is empty.
var ErrorContentTypeIsEmpty = errors.New("content type is empty")

// contentTypes 是一个字符串切片，表示 HTTP 请求的支持内容类型。
// contentTypes is a slice of strings that represents the supported content types for HTTP requests.
var contentTypes = []string{
	com.HttpHeaderJSONContentTypeValue,       // JSON content type
	com.HttpHeaderJavascriptContentTypeValue, // JavaScript content type
	com.HttpHeaderTextContentTypeValue,       // Text content type
	com.HttpHeaderXMLContentTypeValue,        // XML content type
	com.HttpHeaderPXMLContentTypeValue,       // Test XML content type
	com.HttpHeaderYAMLContentTypeValue,       // YAML content type
	com.HttpHeaderTOMLContentTypeValue,       // TOML content type
}

// CalcRequestSize 返回请求对象的大小。
// CalcRequestSize returns the size of the request object
func CalcRequestSize(request *http.Request) int64 {
	size := 0

	// 计算 URL 字符串的长度
	// Calculate the length of the URL string
	if request.URL != nil {
		size += len(request.URL.String())
	}

	// 将请求方法和协议的大小添加到大小变量中
	// Add the size of the request method and protocol to the size variable
	size += len(request.Method)
	size += len(request.Proto)

	// 遍历报头，计算键值对的大小，并将其添加到请求大小中
	// Iterate through the headers, calculate the size of the key-value pairs, and add it to the request size
	for name, values := range request.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}

	// 计算主机名 (Host) 的大小，并将其添加到大小中
	// Add the size of the host name (Host) to the size
	size += len(request.Host)

	// 如果 ContentLength 不是 -1，则将 ContentLength 添加到大小中
	// If ContentLength is not -1, add ContentLength to the size
	if request.ContentLength != -1 {
		size += int(request.ContentLength)
	}

	// 返回请求大小
	// Return the size of the request
	return int64(size)
}

// StringFilterFlags 返回给定字符串中的第一个标记。
// StringFilterFlags returns the first token in the given string
func StringFilterFlags(content string) string {
	// 返回字符串中第一个 ';' 或 ' ' 之前的所有字符。如果两者都不存在，则返回整个字符串。
	// Return all characters before the first ';' or ' ' in the string. If neither exists, return the entire string.
	if i := strings.IndexAny(content, "; "); i >= 0 {
		return content[:i]
	}
	return content
}

// CanRecordContextBody 检查 HTTP 请求头是否包含特定内容类型的值。
// CanRecordContextBody checks if the HTTP request header contains a value for a specific content type
func CanRecordContextBody(header http.Header) bool {
	// 获取请求头中的内容类型
	// Get the content type in the request header
	contentType := header.Get(com.HttpHeaderContentType)

	// 如果请求头为空或内容信息不足以区分类型，则直接返回 false
	// If the request header is empty or the content information is not sufficient to differentiate the type, return false directly
	if contentType == "" || !strings.Contains(contentType, "/") {
		return false
	}

	// 查找定义的内容类型列表中的指定内容类型
	// Find the specified content type in the definedContentTypes list
	typeStr := StringFilterFlags(contentType)
	for _, ct := range contentTypes {
		if strings.HasPrefix(typeStr, ct) {
			return true
		}
	}

	// 如果未定义内容类型，则返回 false
	// Return false if the content type is not defined
	return false
}

// GenerateRequestPath 从 Gin 上下文返回请求路径。
// GenerateRequestPath returns the request path from the Gin context
func GenerateRequestPath(context *gin.Context) string {
	// 如果请求包含查询字符串，则返回整个 URL，否则返回路径
	// If the request contains a query string, return the entire URL, otherwise return the path
	if len(context.Request.URL.RawQuery) > 0 {
		return context.Request.URL.RequestURI()
	}
	return context.Request.URL.Path
}

// GenerateRequestBody 从 Gin 上下文读取 HTTP 请求体并将其存储在缓冲池对象中。
// 请不要直接从 Gin 上下文中读取请求体，因为请求体只能读取一次。
// GenerateRequestBody reads the HTTP request body from the Gin context and stores it in a Buffer Pool object
// Please don't directly to read the request body from the Gin context, because the request body can only be read once
func GenerateRequestBody(context *gin.Context) ([]byte, error) {
	// 检查请求体是否为空
	// Check if the request body is nil
	if context.Request.Body == nil {
		return conver.StringToBytes("request body is nil"), nil
	}

	// 检查是否已经存在相关的 Buffer Pool 对象，如果没有，则创建一个新实例
	// Check if there is already a related Buffer Pool object, if not, create a new instance
	var reqBodyBuffer *bytes.Buffer
	if buffer, ok := context.Get(com.RequestBodyBufferKey); ok {
		reqBodyBuffer = buffer.(*bytes.Buffer)
	} else {
		reqBodyBuffer = com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
	}

	// 检查 Buffer Pool 对象是否为空，如果是，则读取 HTTP 请求体
	// Check if the Buffer Pool object is empty, if so, read the HTTP request body
	if reqBodyBuffer.Len() <= 0 {
		// 读取 HTTP 请求体
		// Read the HTTP request body
		body, err := io.ReadAll(context.Request.Body)
		if err != nil {
			return conver.StringToBytes("failed to get request body"), err
		}

		// 检查是否已经存在相关的 Buffer Pool 对象，如果没有，则创建一个新实例
		// Write the content to the Buffer Pool object
		_, err = reqBodyBuffer.Write(body)
		if err != nil {
			// 如果在将内容写入 Buffer Pool 时发生错误，则存储原始内容
			// If an error occurs while writing the content to the Buff Pool, store the original content
			context.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		} else {
			// 如果内容写入成功，则将内容存储在 Buffer Pool 对象中
			// If the content is written successfully, store the content in the Buffer Pool object
			context.Request.Body = io.NopCloser(reqBodyBuffer)
		}
	}

	// 返回请求体
	// Return the request body
	return reqBodyBuffer.Bytes(), nil
}

// ParseRequestBody 解析请求体为指定类型的值，emptyRequestBodyContent 表示是否允许空的请求体内容。
// ParseRequestBody parses the request body into a variable of the specified type value, emptyRequestBodyContent indicates whether an empty body is allowed
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 检查 ContentType 是否为空
	// Check if ContentType is empty
	if context.ContentType() == "" {
		return ErrorContentTypeIsEmpty
	}

	// 检查请求体是否为空
	// Bind the request body to the specified type value
	var body []byte
	err := context.ShouldBind(value)
	if err != nil {
		// 如果绑定失败，则获取请求体
		// Get the request body
		body, err = GenerateRequestBody(context)
		if err == nil {
			// 如果请求体为空且 emptyRequestBodyContent 为 true，则直接返回 nil
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
