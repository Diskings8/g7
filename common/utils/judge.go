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

	b1 := d2.Equal(d1)
	return b1
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

func getLogicWeek(t time.Time, refreshHour int) (int, int) {
	// 如果当前时间小于当天刷新点，归到前一天
	if t.Hour() < refreshHour {
		t = t.AddDate(0, 0, -1)
	}
	year, week := t.ISOWeek()
	return year, week
}

func CheckTwoTimeIsSameWeek(srcT time.Time, dstT time.Time) bool {
	// 统一转成当地时间（或UTC，看你项目）
	t1 := srcT.Local()
	t2 := dstT.Local()

	// 获取两个时间所在的逻辑周（年+周数）
	d1 := getLogicDay(t1, GameRefreshHour)
	d2 := getLogicDay(t2, GameRefreshHour)
	year1, week1 := d1.ISOWeek()
	year2, week2 := d2.ISOWeek()
	return year1 == year2 && week1 == week2
}

func CheckTwoTimeIsSameMonth(srcT time.Time, dstT time.Time) bool {
	// 统一转成当地时间（或UTC，看你项目）
	t1 := srcT.Local()
	t2 := dstT.Local()

	// 计算两个时间对应的【逻辑天数】
	d1 := getLogicDay(t1, GameRefreshHour)
	d2 := getLogicDay(t2, GameRefreshHour)
	b1 := d2.Year() == d1.Year()
	b2 := d2.Month() == d1.Month()
	return b1 && b2
}
