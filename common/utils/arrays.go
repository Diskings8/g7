package utils

// RemoveElement T 支持任意可比较类型（int/string/struct等）
func RemoveElement[T comparable](slice []T, target T) []T {
	// 遍历查找目标索引
	for i, v := range slice {
		if v == target {
			// 移除：拼接前后，避免内存泄漏
			return removeAt(slice, i)
		}
	}
	// 没找到，返回原切片
	return slice
}

// RemoveAllElement 移除所有匹配的元素
func RemoveAllElement[T comparable](slice []T, target T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != target {
			result = append(result, v)
		}
	}
	return result
}

func removeAt[T any](slice []T, index int) []T {
	// 防止内存泄漏：将最后一个元素置空
	if index < len(slice)-1 {
		slice[index] = slice[len(slice)-1]
		slice[len(slice)-1] = *new(T) // 置零
	}
	return slice[:len(slice)-1]
}
