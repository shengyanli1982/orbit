package conver

import "unsafe"

// BytesToString 字节转字符串
// BytesToString converts byte slice to string
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
