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

	size := int64(len(request.Method) + len(request.Proto) + len(request.Host))

	// 计算 URL 大小
	if url := request.URL; url != nil {
		size += int64(len(url.Scheme) + len(url.Host) + len(url.Path) +
			len(url.RawQuery) + len(url.Fragment))
	}

	// 优化 header 大小计算
	if headers := request.Header; len(headers) > 0 {
		headerSize := 0
		for name, values := range headers {
			headerSize += len(name)
			for _, value := range values {
				headerSize += len(value) + 2 // 2 for ": " separator
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

	// 尝试从上下文获取已存在的缓冲区
	var reqBodyBuffer *bytes.Buffer
	if buffer, exists := context.Get(com.RequestBodyBufferKey); exists {
		if buf, ok := buffer.(*bytes.Buffer); ok {
			if buf.Len() > 0 {
				return buf.Bytes(), nil
			}
			reqBodyBuffer = buf
			reqBodyBuffer.Reset()
		}
	}

	// 如果没有现有缓冲区，从池中获取一个
	if reqBodyBuffer == nil {
		reqBodyBuffer = bufferPool.Get().(*bytes.Buffer)
		reqBodyBuffer.Reset()
		defer func() {
			context.Set(com.RequestBodyBufferKey, reqBodyBuffer)
		}()
	}

	// 读取请求体
	_, err := io.Copy(reqBodyBuffer, context.Request.Body)
	if err != nil {
		return conver.StringToBytes("failed to get request body"), err
	}

	// 重置请求体以供后续读取
	bodyData := reqBodyBuffer.Bytes()
	context.Request.Body = io.NopCloser(bytes.NewReader(bodyData))

	return bodyData, nil
}

// ParseRequestBody 解析请求体。
// ParseRequestBody parses the request body.
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 快速参数验证
	if value == nil {
		return ErrorValueIsNil
	}

	if context == nil {
		return ErrorContextIsNil
	}

	// 获取并验证内容类型（使用 Header.Get 直接获取，避免额外的字符串处理）
	contentType := context.Request.Header.Get(com.HttpHeaderContentType)
	if contentType == "" {
		return ErrorContentTypeIsEmpty
	}

	// 尝试直接绑定请求体到目标结构
	if err := context.ShouldBind(value); err == nil {
		return nil
	}

	// 仅在需要处理空请求体时执行以下逻辑
	if !emptyRequestBodyContent {
		return ErrorBindRequestBody
	}

	// 检查请求体是否为空
	if context.Request.ContentLength == 0 {
		return nil
	}

	// 生成请求体（仅在必要时）
	body, err := GenerateRequestBody(context)
	if err != nil {
		return ErrorGenerateBody
	}

	// 如果请求体为空，返回成功
	if len(body) == 0 {
		return nil
	}

	return ErrorBindRequestBody
}
