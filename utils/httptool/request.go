package httptool

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

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

// 使用 sync.Pool 复用缓冲区
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096)) // 预分配 4KB 缓冲区
	},
}

// CalcRequestSize 计算HTTP请求的总大小（以字节为单位）。
// CalcRequestSize calculates the total size of an HTTP request in bytes.
func CalcRequestSize(request *http.Request) int64 {
	if request == nil {
		return 0
	}

	var size int64
	url := request.URL
	if url != nil {
		size += int64(len(url.Scheme) + len(url.Host) + len(url.Path) +
			len(url.RawQuery) + len(url.Fragment))
	}

	size += int64(len(request.Method) + len(request.Proto) + len(request.Host))

	// 优化 header 大小计算
	if headers := request.Header; len(headers) > 0 {
		var headerSize int
		for name, values := range headers {
			headerSize += len(name)
			for _, value := range values {
				headerSize += len(value) + 2
			}
		}
		size += int64(headerSize)
	}

	if cl := request.ContentLength; cl > 0 {
		size += cl
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
	if contentType == "" || !strings.Contains(contentType, "/") {
		return false
	}

	// 使用 strings.HasPrefix 进行更快的前缀匹配
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
	if context.Request.Body == nil {
		return conver.StringToBytes("request body is nil"), nil
	}

	// 获取或创建缓冲区
	var reqBodyBuffer *bytes.Buffer
	if buffer, exists := context.Get(com.RequestBodyBufferKey); exists {
		if buf, ok := buffer.(*bytes.Buffer); ok {
			reqBodyBuffer = buf
		} else {
			reqBodyBuffer = bufferPool.Get().(*bytes.Buffer)
			context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		}
	} else {
		reqBodyBuffer = bufferPool.Get().(*bytes.Buffer)
		context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
	}

	// 如果缓冲区已有数据，直接返回
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
		context.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		return body, nil
	}

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
