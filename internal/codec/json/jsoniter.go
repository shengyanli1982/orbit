//go:build jsoniter
// +build jsoniter

package json

import jsoniter "github.com/json-iterator/go"

// json 是 jsoniter.ConfigCompatibleWithStandardLibrary 的一个实例
var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)
