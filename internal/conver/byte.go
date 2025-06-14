package conver

import "unsafe"

// 将字节切片转换为字符串的函数
func BytesToString(b []byte) string {
	// 快速路径：空切片直接返回空字符串
	if len(b) == 0 {
		return ""
	}

	return *(*string)(unsafe.Pointer(&b))
}
