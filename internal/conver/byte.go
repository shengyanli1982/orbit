package conver

import "unsafe"

// BytesToString 是一个将字节切片转换为字符串的函数。
// BytesToString is a function that converts a byte slice to a string.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
