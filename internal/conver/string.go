package conver

import "unsafe"

// StringToBytes 是一个将字符串转换为字节切片的函数。
// StringToBytes is a function that converts a string to a byte slice.
func StringToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
