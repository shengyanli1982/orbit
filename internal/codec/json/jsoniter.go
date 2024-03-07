//go:build jsoniter
// +build jsoniter

package json

import jsoniter "github.com/json-iterator/go"

// json 是 jsoniter.ConfigCompatibleWithStandardLibrary 的一个实例。
// json is an instance of jsoniter.ConfigCompatibleWithStandardLibrary.
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal 是一个将 Go 值转换为 JSON 的函数。
// Marshal is a function that converts a Go value to JSON.
var Marshal = json.Marshal

// Unmarshal 是一个将 JSON 转换为 Go 值的函数。
// Unmarshal is a function that converts JSON to a Go value.
var Unmarshal = json.Unmarshal

// MarshalIndent 是一个将 Go 值转换为带缩进的 JSON 的函数。
// MarshalIndent is a function that converts a Go value to JSON with indentation.
var MarshalIndent = json.MarshalIndent

// NewDecoder 是一个创建新的 JSON 解码器的函数。
// NewDecoder is a function that creates a new JSON decoder.
var NewDecoder = json.NewDecoder

// NewEncoder 是一个创建新的 JSON 编码器的函数。
// NewEncoder is a function that creates a new JSON encoder.
var NewEncoder = json.NewEncoder

// RawMessage 是一个原始编码的 JSON 值。
// RawMessage is a raw encoded JSON value.
type RawMessage = json.RawMessage
