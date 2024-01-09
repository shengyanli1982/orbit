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
	ErrorContentTypeIsEmpty = errors.New("content type is empty")
)

var contentTypes = []string{
	com.HttpHeaderJSONContentTypeValue,
	com.HttpHeaderJavascriptContentTypeValue,
	com.HttpHeaderTextContentTypeValue,
	com.HttpHeaderXMLContentTypeValue,
	com.HttpHeaderXML2ContentTypeValue,
	com.HttpHeaderYAMLContentTypeValue,
	com.HttpHeaderTOMLContentTypeValue,
}

// CalcRequestSize返回请求(request)对象的大小
func CalcRequestSize(r *http.Request) int64 {
	size := 0

	// 计算URL的字符串长度
	if r.URL != nil {
		size += len(r.URL.String())
	}

	// 将方法(Method)和协议(Proto)放到size变量中
	size += len(r.Method)
	size += len(r.Proto)

	// 遍历 header，统计 header 的键值对大小，并加到请求大小中
	for name, values := range r.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}

	// size中增加了主机名(Host)的大小
	size += len(r.Host)

	// 如果ContentLength设置不为-1，则将ContentLength添加到size中
	if r.ContentLength != -1 {
		size += int(r.ContentLength)
	}

	return int64(size)
}

// StringFilterFlags 返回给定字符串中第一个标记。
func StringFilterFlags(content string) string {
	// 将字符串中第一个 ';' 或 ' ' 前的所有字符作为第一个标记。如果两者都不存在，则返回整个字符串。
	if i := strings.IndexAny(content, "; "); i >= 0 {
		return content[:i]
	}
	return content
}

// CanRecordContextBody 检查HTTP请求头中是否存在特定内容类型的值。
func CanRecordContextBody(h http.Header) bool {
	v := h.Get(com.HttpHeaderContentType)

	// 如果请求头为空或者内容信息不足以区分类型，则直接返回false
	if v == "" || !strings.Contains(v, "/") {
		return false
	}

	// 查找所有 definedContentTypes 列表中指定的内容类型。
	typeStr := StringFilterFlags(v)
	for _, ct := range contentTypes {
		if strings.HasPrefix(typeStr, ct) {
			return true
		}
	}

	// 如果内容类型未被定义则返回false
	return false
}

func GenerateRequestPath(c *gin.Context) string {
	if len(c.Request.URL.RawQuery) > 0 {
		return c.Request.URL.RequestURI()
	}
	return c.Request.URL.Path
}

// GenerateRequestBody 从Gin的上下文中读取HTTP请求的Body，并将其存储到一个Buffer Pool对象中。
func GenerateRequestBody(c *gin.Context) ([]byte, error) {
	// 检查是否已经有相关Buffer Pool对象，如果没有，则创建一个新的实例
	var buf *bytes.Buffer
	if o, ok := c.Get(com.RequestBodyBufferKey); ok {
		buf = o.(*bytes.Buffer)
	} else {
		buf = com.RequestBodyBufferPool.Get()
		c.Set(com.RequestBodyBufferKey, buf)
	}

	// 如果Buffer Pool对象已经被使用过，则需要先清空
	buf.Reset()

	// 读取HTTP请求的Body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return conver.StringToBytes("failed to get request body"), err
	}

	// 把内容写入 Buffer Pool 对象中
	_, err = buf.Write(body)
	if err != nil {
		// 如果在把内容写入 Buff Pool 是出现了错误，则存储原始的内容
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	} else {
		c.Request.Body = io.NopCloser(buf)
	}

	// 返回请求的 Body
	return body, nil
}

// ParseRequestBody 将 request body 解析为指定类型 value 的变量，emptyRequestBodyContent 表示是否允许为空。
func ParseRequestBody(c *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// 判断 ContentType 是否为空
	if c.ContentType() == "" {
		return ErrorContentTypeIsEmpty
	}

	var b []byte
	err := c.ShouldBind(value)
	if err != nil {
		b, err = GenerateRequestBody(c)
		if err == nil {
			if emptyRequestBodyContent && len(b) <= 0 {
				return nil
			}
		}

	}

	return err
}
