package conver

import "unsafe"

// BytesToString 将字节切片转换为字符串。
// BytesToString converts a byte slice to a string.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
