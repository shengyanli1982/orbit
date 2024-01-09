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
	ErrContentTypeIsEmpty = errors.New("content type is empty")
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

// CalcRequestSize returns the size of the request object
func CalcRequestSize(request *http.Request) int64 {
	size := 0

	// Calculate the length of the URL string
	if request.URL != nil {
		size += len(request.URL.String())
	}

	// Add the method and protocol to the size variable
	size += len(request.Method)
	size += len(request.Proto)

	// Iterate through the headers, calculate the size of the key-value pairs, and add it to the request size
	for name, values := range request.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}

	// Add the size of the host name (Host) to the size
	size += len(request.Host)

	// If ContentLength is not -1, add ContentLength to the size
	if request.ContentLength != -1 {
		size += int(request.ContentLength)
	}

	return int64(size)
}

// StringFilterFlags returns the first token in the given string
func StringFilterFlags(content string) string {
	// Return all characters before the first ';' or ' ' in the string. If neither exists, return the entire string.
	if i := strings.IndexAny(content, "; "); i >= 0 {
		return content[:i]
	}
	return content
}

// CanRecordContextBody checks if the HTTP request header contains a value for a specific content type
func CanRecordContextBody(header http.Header) bool {
	contentType := header.Get(com.HttpHeaderContentType)

	// If the request header is empty or the content information is not sufficient to differentiate the type, return false directly
	if contentType == "" || !strings.Contains(contentType, "/") {
		return false
	}

	// Find the specified content type in the definedContentTypes list
	typeStr := StringFilterFlags(contentType)
	for _, ct := range contentTypes {
		if strings.HasPrefix(typeStr, ct) {
			return true
		}
	}

	// Return false if the content type is not defined
	return false
}

func GenerateRequestPath(context *gin.Context) string {
	if len(context.Request.URL.RawQuery) > 0 {
		return context.Request.URL.RequestURI()
	}
	return context.Request.URL.Path
}

// GenerateRequestBody reads the HTTP request body from the Gin context and stores it in a Buffer Pool object
func GenerateRequestBody(context *gin.Context) ([]byte, error) {
	// Check if there is already a related Buffer Pool object, if not, create a new instance
	var buf *bytes.Buffer
	if obj, ok := context.Get(com.RequestBodyBufferKey); ok {
		buf = obj.(*bytes.Buffer)
	} else {
		buf = com.RequestBodyBufferPool.Get()
		context.Set(com.RequestBodyBufferKey, buf)
	}

	// Reset the Buffer Pool object if it has been used before
	buf.Reset()

	// Read the HTTP request body
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		return conver.StringToBytes("failed to get request body"), err
	}

	// Write the content to the Buffer Pool object
	_, err = buf.Write(body)
	if err != nil {
		// If an error occurs while writing the content to the Buff Pool, store the original content
		context.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	} else {
		context.Request.Body = io.NopCloser(buf)
	}

	// Return the request body
	return body, nil
}

// ParseRequestBody parses the request body into a variable of the specified type value, emptyRequestBodyContent indicates whether an empty body is allowed
func ParseRequestBody(context *gin.Context, value interface{}, emptyRequestBodyContent bool) error {
	// Check if ContentType is empty
	if context.ContentType() == "" {
		return ErrContentTypeIsEmpty
	}

	var body []byte
	err := context.ShouldBind(value)
	if err != nil {
		body, err = GenerateRequestBody(context)
		if err == nil {
			if emptyRequestBodyContent && len(body) <= 0 {
				return nil
			}
		}

	}

	return err
}
