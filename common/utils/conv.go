package utils

import (
	"errors"
	"fmt"
	"strconv"
)

func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Int32ToUint8(i int32) (uint8, error) {
	// 必须判断范围！
	if i < 0 || i > 255 {
		return 0, errors.New(fmt.Sprintf("%d too big to change", i))
	}
	return uint8(i), nil
}
