package conver

import "unsafe"

// BytesToString converts a byte slice to a string.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
