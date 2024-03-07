package conver

import "unsafe"

// StringToBytes 是一个将字符串转换为字节切片的函数。
// StringToBytes is a function that converts a string to a byte slice.
func StringToBytes(s string) []byte {
	// 使用 unsafe.Pointer 将字符串的地址转换为 uintptr 数组的指针。
	// Use unsafe.Pointer to convert the address of the string to a pointer to a uintptr array.
	x := (*[2]uintptr)(unsafe.Pointer(&s))

	// 创建一个新的 uintptr 数组，其中包含字符串的起始地址、结束地址和结束地址。
	// Create a new uintptr array that contains the start address, end address, and end address of the string.
	h := [3]uintptr{x[0], x[1], x[1]}

	// 使用 unsafe.Pointer 将 uintptr 数组的地址转换为字节切片的指针，然后解引用得到字节切片。
	// Use unsafe.Pointer to convert the address of the uintptr array to a pointer to a byte slice, then dereference to get the byte slice.
	return *(*[]byte)(unsafe.Pointer(&h))
}
