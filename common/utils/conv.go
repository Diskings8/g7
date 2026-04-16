package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Int32ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

func StringToInit64(i string) int64 {
	num, _ := strconv.ParseInt(i, 10, 64)
	return num
}

func StringToInit32(i string) int32 {
	num, _ := strconv.ParseInt(i, 10, 32)
	return int32(num)
}

func Int32ToUint8(i int32) (uint8, error) {
	// 必须判断范围！
	if i < 0 || i > 255 {
		return 0, errors.New(fmt.Sprintf("%d too big to change", i))
	}
	return uint8(i), nil
}

func FormatTimestamp(t int64) time.Time {
	return time.Unix(t, 0).UTC()
}

func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}
