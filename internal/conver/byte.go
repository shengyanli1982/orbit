package conver

import "unsafe"

// BytesToString 将字节切片转换为字符串
// 使用 unsafe.Pointer 进行零拷贝转换，提高性能
func BytesToString(b []byte) string {
	// 快速路径：空切片直接返回空字符串
	if len(b) == 0 {
		return ""
	}

	return *(*string)(unsafe.Pointer(&b))
}
