package structs

import "fmt"

type KInt32VInt64 struct {
	K int32
	V int64
}

type KVString struct {
	K string
	V string
}

type KInt32VInt64Bind struct {
	K int32
	V int64
	B int32
}

func MergeKInt32VInt64Bind(src []KInt32VInt64Bind) []KInt32VInt64Bind {
	if len(src) == 0 {
		return nil
	}

	// 用 map 做聚合：key = "K:B"，value = 合并后的结构
	temp := make(map[string]*KInt32VInt64Bind)

	for _, item := range src {
		key := fmt.Sprintf("%d:%d", item.K, item.B)
		if existing, ok := temp[key]; ok {
			// 数量累加
			existing.V += item.V
		} else {
			// 新建
			temp[key] = &KInt32VInt64Bind{
				K: item.K,
				V: item.V,
				B: item.B,
			}
		}
	}

	// 转回 slice
	res := make([]KInt32VInt64Bind, 0, len(temp))
	for _, v := range temp {
		res = append(res, *v)
	}

	return res
}
