package conver

import "unsafe"

// 将字符串转换为字节切片的函数
func StringToBytes(s string) []byte {
	// 快速路径：空字符串直接返回空切片
	if len(s) == 0 {
		return []byte{}
	}

	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
