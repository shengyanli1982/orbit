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

var (
	ErrorContextIsNil       = errors.New("context is nil")
	ErrorValueIsNil         = errors.New("value is nil")
	ErrorContentTypeIsEmpty = errors.New("content type is empty")
	ErrorBindRequestBody    = errors.New("failed to bind request body")
	ErrorGenerateBody       = errors.New("failed to generate request body")
)

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
	if request == nil {
		return 0
	}

	// 使用 int64 避免大请求时的整数溢出
	var size int64

	// URL 大小计算
	if url := request.URL; url != nil {
		// 预计算 URL 各部分
		size += int64(len(url.Scheme) +
			len(url.Host) +
			len(url.Path) +
			len(url.RawQuery) +
			len(url.Fragment))
	}

	// 基本请求信息
	size += int64(len(request.Method) +
		len(request.Proto) +
		len(request.Host))

	// 优化 Header 大小计算
	if headers := request.Header; len(headers) > 0 {
		for name, values := range headers {
			headerSize := len(name)
			// 预估每个值的大小并包含分隔符的长度
			for _, value := range values {
				headerSize += len(value) + 2 // 2 for ": " or ", "
			}
			size += int64(headerSize)
		}
	}

	// Content-Length 处理
	if cl := request.ContentLength; cl > 0 {
		size += cl
	}

	// 处理 Transfer-Encoding
	if te := request.TransferEncoding; len(te) > 0 {
		for _, encoding := range te {
			size += int64(len(encoding) + 2) // 2 for ", "
		}
	}

	// 处理 Trailer
	if trailers := request.Trailer; len(trailers) > 0 {
		for name, values := range trailers {
			size += int64(len(name))
			for _, value := range values {
				size += int64(len(value) + 2) // 2 for ": " or ", "
			}
		}
	}

	return size
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
	contentType := StringFilterFlags(header.Get(com.HttpHeaderContentType))

	// 如果请求头为空或内容信息不足以区分类型，直接返回 false
	// If the request header is empty or the content information is not sufficient to differentiate the type, return false directly
	if contentType == "" || strings.IndexByte(contentType, '/') == -1 {
		return false
	}

	// 在 definedContentTypes 列表中查找指定的内容类型
	// Find the specified content type in the definedContentTypes list
	for _, ct := range contentTypes {
		if strings.HasPrefix(contentType, ct) {
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
	// 快速路径：检查请求体是否为空
	if context.Request.Body == nil {
		return conver.StringToBytes("request body is nil"), nil
	}

	// 获取或创建缓冲区
	var reqBodyBuffer *bytes.Buffer
	if buffer, exists := context.Get(com.RequestBodyBufferKey); exists {
		if buf, ok := buffer.(*bytes.Buffer); ok {
			reqBodyBuffer = buf
		} else {
			// 类型断言失败，创建新缓冲区
			reqBodyBuffer = com.RequestBodyBufferPool.Get()
			context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		}
	} else {
		reqBodyBuffer = com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
	}

	// 如果缓冲区已有内容，直接返回
	if reqBodyBuffer.Len() > 0 {
		return reqBodyBuffer.Bytes(), nil
	}

	// 读取请求体
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		return conver.StringToBytes("failed to get request body"), err
	}

	// 写入缓冲区
	if _, err := reqBodyBuffer.Write(body); err != nil {
		// 写入失败时使用原始数据
		context.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		return body, nil
	}

	// 写入成功，使用缓冲区数据
	context.Request.Body = io.NopCloser(reqBodyBuffer)
	return reqBodyBuffer.Bytes(), nil
}

// ParseRequestBody 函数将请求体解析到指定类型的变量 value 中，emptyRequestBodyContent 表示是否允许请求体为空
// The ParseRequestBody function parses the request body into a variable of the specified type value, emptyRequestBodyContent indicates whether an empty body is allowed
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 1. 参数验证
	if context == nil {
		return ErrorContextIsNil
	}
	if value == nil {
		return ErrorValueIsNil
	}

	// 2. 检查 Content-Type
	contentType := context.ContentType()
	if contentType == "" {
		return ErrorContentTypeIsEmpty
	}

	// 3. 尝试绑定请求体
	if err := context.ShouldBind(value); err == nil {
		return nil
	}

	// 4. 处理空请求体的情况
	if emptyRequestBodyContent {
		body, err := GenerateRequestBody(context)
		if err != nil {
			return ErrorGenerateBody
		}

		if len(body) == 0 {
			return nil
		}
	}

	// 5. 返回绑定错误
	return ErrorBindRequestBody
}
