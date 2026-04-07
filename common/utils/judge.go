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

	return d2.After(d1)
}

func getLogicDay(t time.Time, refreshHour int) time.Time {
	y, m, d := t.Date()

	// 如果当前小时 < 刷新小时 → 算作前一天
	if t.Hour() < refreshHour {
		// 往前推一天
		t = t.Add(-24 * time.Hour)
		y, m, d = t.Date()
	}

	// 返回一个唯一数字代表这一天
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func CheckTwoTimeIsSameWeek(srcT time.Time, dstT time.Time) bool {
	// 统一转成当地时间（或UTC，看你项目）
	t1 := srcT.Local()
	t2 := dstT.Local()

	// 计算两个时间对应的【逻辑天数】
	d1 := getLogicDay(t1, GameRefreshHour)
	d2 := getLogicDay(t2, GameRefreshHour)

	// 现在是周一
	if d1.Weekday() != time.Monday {
		return false
	}
	return d2.After(d1)
}

func CheckTwoTimeIsSameMonth(srcT time.Time, dstT time.Time) bool {
	// 统一转成当地时间（或UTC，看你项目）
	t1 := srcT.Local()
	t2 := dstT.Local()

	// 计算两个时间对应的【逻辑天数】
	d1 := getLogicDay(t1, GameRefreshHour)
	d2 := getLogicDay(t2, GameRefreshHour)

	// 现在是周一
	if d1.Day() != 1 {
		return false
	}
	return d2.Year() != d1.Year() || d2.Month() != d1.Month()
}
