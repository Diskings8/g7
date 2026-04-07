package cronx

import (
	"github.com/robfig/cron/v3"
)

var GCron = &CronAllTask{}

type CronAllTask struct {
	cron *cron.Cron
}

// 初始化定时任务
func InitCron() {
	// 秒级 cron
	GCron.cron = cron.New(cron.WithSeconds())
	GCron.cron.Start()
}

const (
	EveryDay0Hour = "0 0 0 * * ?"
	EveryDay5Hour = "0 0 5 * * ?"
)

func addCronTask(expr string, task func()) (cron.EntryID, error) {
	return GCron.cron.AddFunc(expr, task)
}

func AddDaily5HourTask(task func()) {
	_, _ = addCronTask(EveryDay5Hour, task)
}

func AddDaily0HourTask(task func()) {
	_, _ = addCronTask(EveryDay0Hour, task)
}
