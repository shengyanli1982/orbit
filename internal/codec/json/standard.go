//go:build !jsoniter
// +build !jsoniter

package json

import "encoding/json"

// Marshal 是 json.Marshal 的别名。
// Marshal is a function that converts a Go value to JSON.
var Marshal = json.Marshal

// Unmarshal 是 json.Unmarshal 的别名。
// Unmarshal is a function that converts JSON to a Go value.
var Unmarshal = json.Unmarshal

// MarshalIndent 是 json.MarshalIndent 的别名。
// MarshalIndent is a function that converts a Go value to JSON with indentation.
var MarshalIndent = json.MarshalIndent

// NewDecoder 是 json.NewDecoder 的别名。
// NewDecoder is a function that creates a new JSON decoder.
var NewDecoder = json.NewDecoder

// NewEncoder 是 json.NewEncoder 的别名。
// NewEncoder is a function that creates a new JSON encoder.
var NewEncoder = json.NewEncoder

// RawMessage 是 json.RawMessage 的别名。
// RawMessage is a raw encoded JSON value.
type RawMessage = json.RawMessage
