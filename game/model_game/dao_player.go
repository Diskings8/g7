package model_game

import (
	"encoding/json"
	"time"
)

type PlayerDao struct {
	PlayerId         int64           `gorm:"primaryKey column:user_id"`
	UserId           int64           `gorm:"column:user_id"`
	ServerId         int32           `gorm:"column:server_id"`
	Nickname         string          `gorm:"column:nickname"`
	IsOnline         bool            `gorm:"column:is_online"`          // 是否在线
	OfflineAt        time.Time       `gorm:"column:offline_at"`         // 离线时间
	OnlineAt         time.Time       `gorm:"column:online_at"`          // 当前上线时间
	LastOfflineAt    time.Time       `gorm:"column:last_offline_at"`    // 上次离线时间
	LastDailyResetAt time.Time       `gorm:"column:last_dailyreset_at"` // 上次每日重置的时间
	LastWeekResetAt  time.Time       `gorm:"column:last_weekreset_at"`  // 每周重置时间
	LastMonthResetAt time.Time       `gorm:"column:last_monthreset_at"` // 每月重置时间
	GeneralD         generalData     `gorm:"-"`
	generalData      []byte          `gorm:"column:general_data"`
	CultivationD     cultivationData `gorm:"-"`
	cultivationData  []byte          `gorm:"column:cultivation_data"`
	ActivityD        activityData    `gorm:"-"`
	activityData     []byte          `gorm:"column:activity_data"`
}

type generalData struct {
	BagData []byte `json:"bag_data"`
}

type cultivationData struct {
}

type activityData struct {
}

func (dao *PlayerDao) Unmarshal() {
	_ = json.Unmarshal(dao.generalData, &dao.GeneralD)
	_ = json.Unmarshal(dao.cultivationData, &dao.CultivationD)
	_ = json.Unmarshal(dao.activityData, &dao.ActivityD)
}

func (dao *PlayerDao) TableName() string {
	return "player_dao"
}

func (this *PlayerDao) TomSimplePlayer() *Player {
	p := &Player{
		PlayerId:      this.PlayerId,
		UserId:        this.UserId,
		ServerId:      this.ServerId,
		Nickname:      this.Nickname,
		IsOnline:      this.IsOnline,
		OfflineAt:     this.OfflineAt,
		OnlineAt:      this.OnlineAt,
		LastOfflineAt: this.LastOfflineAt,
	}
	return p
}
