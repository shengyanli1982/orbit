//go:build !jsoniter
// +build !jsoniter

package json

import "encoding/json"

// Marshal is a function that converts a Go value to JSON.
var Marshal = json.Marshal

// Unmarshal is a function that converts JSON to a Go value.
var Unmarshal = json.Unmarshal

// MarshalIndent is a function that converts a Go value to JSON with indentation.
var MarshalIndent = json.MarshalIndent

// NewDecoder is a function that creates a new JSON decoder.
var NewDecoder = json.NewDecoder

// NewEncoder is a function that creates a new JSON encoder.
var NewEncoder = json.NewEncoder

// RawMessage is a raw encoded JSON value.
type RawMessage = json.RawMessage
