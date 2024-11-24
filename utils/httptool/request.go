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

// 定义错误变量。
// Define error variables.
var (
	ErrorContextIsNil       = errors.New("context is nil")                  // 上下文为空错误 (Context is nil error)
	ErrorValueIsNil         = errors.New("value is nil")                    // 值为空错误 (Value is nil error)
	ErrorContentTypeIsEmpty = errors.New("content type is empty")           // 内容类型为空错误 (Content type is empty error)
	ErrorBindRequestBody    = errors.New("failed to bind request body")     // 绑定请求体失败错误 (Failed to bind request body error)
	ErrorGenerateBody       = errors.New("failed to generate request body") // 生成请求体失败错误 (Failed to generate request body error)
)

// contentTypes 包含支持的内容类型列表。
// contentTypes contains a list of supported content types.
var contentTypes = []string{
	com.HttpHeaderJSONContentTypeValue,       // JSON内容类型 (JSON content type)
	com.HttpHeaderJavascriptContentTypeValue, // JavaScript内容类型 (JavaScript content type)
	com.HttpHeaderTextContentTypeValue,       // 文本内容类型 (Text content type)
	com.HttpHeaderXMLContentTypeValue,        // XML内容类型 (XML content type)
	com.HttpHeaderPXMLContentTypeValue,       // PXML内容类型 (PXML content type)
	com.HttpHeaderYAMLContentTypeValue,       // YAML内容类型 (YAML content type)
	com.HttpHeaderTOMLContentTypeValue,       // TOML内容类型 (TOML content type)
}

// CalcRequestSize 计算HTTP请求的总大小（以字节为单位）。
// CalcRequestSize calculates the total size of an HTTP request in bytes.
func CalcRequestSize(request *http.Request) int64 {
	if request == nil {
		return 0
	}

	var size int64

	// 计算URL各部分的大小
	// Calculate the size of URL components
	if url := request.URL; url != nil {
		size += int64(len(url.Scheme) +
			len(url.Host) +
			len(url.Path) +
			len(url.RawQuery) +
			len(url.Fragment))
	}

	// 计算请求基本信息的大小
	// Calculate the size of basic request information
	size += int64(len(request.Method) +
		len(request.Proto) +
		len(request.Host))

	// 计算请求头的大小
	// Calculate the size of request headers
	if headers := request.Header; len(headers) > 0 {
		for name, values := range headers {
			headerSize := len(name)
			for _, value := range values {
				headerSize += len(value) + 2 // 加2是为了包含": "分隔符 (Add 2 for ": " separator)
			}
			size += int64(headerSize)
		}
	}

	// 添加内容长度
	// Add content length
	if cl := request.ContentLength; cl > 0 {
		size += cl
	}

	// 计算传输编码的大小
	// Calculate the size of transfer encoding
	if te := request.TransferEncoding; len(te) > 0 {
		for _, encoding := range te {
			size += int64(len(encoding) + 2) // 加2是为了包含分隔符 (Add 2 for separator)
		}
	}

	// 计算尾部头信息的大小
	// Calculate the size of trailer headers
	if trailers := request.Trailer; len(trailers) > 0 {
		for name, values := range trailers {
			size += int64(len(name))
			for _, value := range values {
				size += int64(len(value) + 2) // 加2是为了包含": "分隔符 (Add 2 for ": " separator)
			}
		}
	}

	return size
}

// StringFilterFlags 从内容类型字符串中过滤掉标志。
// StringFilterFlags filters out flags from the content type string.
func StringFilterFlags(content string) string {
	if i := strings.IndexAny(content, "; "); i >= 0 {
		return content[:i]
	}
	return content
}

// CanRecordContextBody 检查是否可以记录请求体。
// CanRecordContextBody checks if the request body can be recorded.
func CanRecordContextBody(header http.Header) bool {
	contentType := StringFilterFlags(header.Get(com.HttpHeaderContentType))

	// 检查内容类型是否为空或无效
	// Check if content type is empty or invalid
	if contentType == "" || strings.IndexByte(contentType, '/') == -1 {
		return false
	}

	// 检查是否为支持的内容类型
	// Check if it's a supported content type
	for _, ct := range contentTypes {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}

	return false
}

// GenerateRequestPath 生成请求路径。
// GenerateRequestPath generates the request path.
func GenerateRequestPath(context *gin.Context) string {
	if len(context.Request.URL.RawQuery) > 0 {
		return context.Request.URL.RequestURI()
	}
	return context.Request.URL.Path
}

// GenerateRequestBody 生成请求体。
// GenerateRequestBody generates the request body.
func GenerateRequestBody(context *gin.Context) ([]byte, error) {
	// 检查请求体是否为空
	// Check if request body is nil
	if context.Request.Body == nil {
		return conver.StringToBytes("request body is nil"), nil
	}

	var reqBodyBuffer *bytes.Buffer
	// 尝试从上下文中获取已存在的缓冲区
	// Try to get existing buffer from context
	if buffer, exists := context.Get(com.RequestBodyBufferKey); exists {
		// 如果缓冲区类型正确，直接使用
		// If buffer type is correct, use it directly
		if buf, ok := buffer.(*bytes.Buffer); ok {
			reqBodyBuffer = buf
		} else {
			// 类型不正确，创建新的缓冲区
			// If type is incorrect, create new buffer
			reqBodyBuffer = com.RequestBodyBufferPool.Get()
			context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		}
	} else {
		// 上下文中不存在缓冲区，创建新的
		// Buffer doesn't exist in context, create new one
		reqBodyBuffer = com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
	}

	// 如果缓冲区已有数据，直接返回
	// If buffer already has data, return it directly
	if reqBodyBuffer.Len() > 0 {
		return reqBodyBuffer.Bytes(), nil
	}

	// 读取请求体内容
	// Read request body content
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		return conver.StringToBytes("failed to get request body"), err
	}

	// 尝试将内容写入缓冲区
	// Try to write content to buffer
	if _, err := reqBodyBuffer.Write(body); err != nil {
		// 写入失败时，使用原始body数据
		// If write fails, use original body data
		context.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		return body, nil
	}

	// 重新设置请求体，使其可以被后续中间件读取
	// Reset request body so it can be read by subsequent middleware
	context.Request.Body = io.NopCloser(reqBodyBuffer)
	return reqBodyBuffer.Bytes(), nil
}

// ParseRequestBody 解析请求体。
// ParseRequestBody parses the request body.
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 验证基本参数
	// Validate basic parameters
	if context == nil {
		return ErrorContextIsNil
	}
	if value == nil {
		return ErrorValueIsNil
	}

	// 获取并验证内容类型
	// Get and validate content type
	contentType := context.ContentType()
	if contentType == "" {
		return ErrorContentTypeIsEmpty
	}

	// 尝试绑定请求体到目标结构
	// Try to bind request body to target structure
	if err := context.ShouldBind(value); err == nil {
		return nil
	}

	// 处理空请求体的情况
	// Handle empty request body case
	if emptyRequestBodyContent {
		// 生成请求体
		// Generate request body
		body, err := GenerateRequestBody(context)
		if err != nil {
			return ErrorGenerateBody
		}

		// 如果请求体为空，返回成功
		// If request body is empty, return success
		if len(body) == 0 {
			return nil
		}
	}

	return ErrorBindRequestBody
}
