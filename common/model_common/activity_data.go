package model_common

import (
	"time"
)

type BaseActivity struct {
	ActivityId   int64  `gorm:"column:activity_id;primaryKey;autoIncrement"`
	ConfId       int32  `gorm:"column:conf_id;"`
	ActivityType int32  `gorm:"column:activity_type;default:0" json:"activity_type"`
	Title        string `gorm:"column:title;type:varchar(128);not null" json:"title"`
	Desc         string `gorm:"column:desc;type:varchar(128);not null" json:"desc"`
	Icon         string `gorm:"column:icon;type:varchar(128);not null" json:"icon"`
	StartTime    int64  `gorm:"column:start_time" json:"start_time"`
	EndTime      int64  `gorm:"column:end_time" json:"end_time"`
	CloseTime    int64  `gorm:"column:close_time" json:"close_time"`
	Status       int32  `gorm:"column:status;default:0;comment:0=不生效 1=已生效" json:"status"`
	ConfData     []byte `gorm:"column:conf_data;type:text" json:"conf_data"`
	DeleteFlag   int32  `gorm:"column:delete_flag;default:0" json:"delete_flag"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (BaseActivity) TableName() string {
	return "base_activity"
}

type GameActivity struct {
	BaseActivity
	ServerId       int32 `gorm:"column:server_id;"`
	BaseActivityId int64 `gorm:"column:base_activity_id;"`
}

func (GameActivity) TableName() string {
	return "game_activity"
}
