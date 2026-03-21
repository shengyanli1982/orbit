//go:build sonic
// +build sonic

package json

import "github.com/bytedance/sonic"

var json = sonic.ConfigStd

var (
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)
