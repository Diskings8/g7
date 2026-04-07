package utils

import "time"

const (
	GameRefreshHour = 0
)

func CheckTwoTimeIsSameDay(srcT time.Time, dstT time.Time) bool {
	// 统一转成当地时间（或UTC，看你项目）
	t1 := srcT.Local()
	t2 := dstT.Local()

	// 计算两个时间对应的【逻辑天数】
	d1 := getLogicDay(t1, GameRefreshHour)
	d2 := getLogicDay(t2, GameRefreshHour)

	return d1 == d2
}

func getLogicDay(t time.Time, refreshHour int) int64 {
	y, m, d := t.Date()

	// 如果当前小时 < 刷新小时 → 算作前一天
	if t.Hour() < refreshHour {
		// 往前推一天
		t = t.Add(-24 * time.Hour)
		y, m, d = t.Date()
	}

	// 返回一个唯一数字代表这一天
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location()).Unix()
}
