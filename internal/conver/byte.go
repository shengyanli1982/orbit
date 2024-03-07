package conver

import "unsafe"

// BytesToString 是一个将字节切片转换为字符串的函数。
// BytesToString is a function that converts a byte slice to a string.
func BytesToString(b []byte) string {
	// 使用 unsafe.Pointer 将字节切片的地址转换为字符串的指针，然后解引用得到字符串。
	// Use unsafe.Pointer to convert the address of the byte slice to a pointer to a string, then dereference to get the string.
	return *(*string)(unsafe.Pointer(&b))
}
