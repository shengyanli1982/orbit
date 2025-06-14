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

// 定义错误变量
var (
	ErrorContextIsNil       = errors.New("context is nil")
	ErrorValueIsNil         = errors.New("value is nil")
	ErrorContentTypeIsEmpty = errors.New("content type is empty")
	ErrorBindRequestBody    = errors.New("failed to bind request body")
	ErrorGenerateBody       = errors.New("failed to generate request body")
)

// contentTypes 包含支持的内容类型列表
var contentTypes = []string{
	com.HttpHeaderJSONContentTypeValue,
	com.HttpHeaderJavascriptContentTypeValue,
	com.HttpHeaderTextContentTypeValue,
	com.HttpHeaderXMLContentTypeValue,
	com.HttpHeaderPXMLContentTypeValue,
	com.HttpHeaderYAMLContentTypeValue,
	com.HttpHeaderTOMLContentTypeValue,
}

// 计算HTTP请求的总大小（以字节为单位）
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

// 从内容类型字符串中过滤掉标志
func StringFilterFlags(content string) string {
	if i := strings.IndexAny(content, "; "); i >= 0 {
		return content[:i]
	}
	return content
}

// 检查是否可以记录请求体
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

// 生成请求路径
func GenerateRequestPath(context *gin.Context) string {
	if len(context.Request.URL.RawQuery) > 0 {
		return context.Request.URL.RequestURI()
	}
	return context.Request.URL.Path
}

// 生成请求体
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
		reqBodyBuffer = com.RequestBodyBufferPool.Get()
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

// 解析请求体
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 快速参数验证
	if value == nil {
		return ErrorValueIsNil
	}

	if context == nil {
		return ErrorContextIsNil
	}

	// 获取并验证内容类型
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
