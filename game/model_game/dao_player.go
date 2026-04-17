package model_game

import (
	"g7/common/utils"
)

type PlayerDao struct {
	PlayerId         int64           `gorm:"primaryKey;column:player_id"`
	UserId           int64           `gorm:"column:user_id"`
	ServerId         int32           `gorm:"column:server_id"`
	Nickname         string          `gorm:"column:nickname"`
	IsOnline         bool            `gorm:"column:is_online"`          // 是否在线
	OfflineAt        int64           `gorm:"column:offline_at"`         // 离线时间
	OnlineAt         int64           `gorm:"column:online_at"`          // 当前上线时间
	LastOfflineAt    int64           `gorm:"column:last_offline_at"`    // 上次离线时间
	LastDailyResetAt int64           `gorm:"column:last_dailyreset_at"` // 上次每日重置的时间
	LastWeekResetAt  int64           `gorm:"column:last_weekreset_at"`  // 每周重置时间
	LastMonthResetAt int64           `gorm:"column:last_monthreset_at"` // 每月重置时间
	GeneralD         generalData     `gorm:"-" json:"-"`
	GeneralData      []byte          `gorm:"column:general_data"`
	CultivationD     cultivationData `gorm:"-" json:"-"`
	CultivationData  []byte          `gorm:"column:cultivation_data"`
	ActivityD        activityData    `gorm:"-" json:"-"`
	ActivityData     []byte          `gorm:"column:activity_data"`
}

type generalData struct {
	BagData  AllBagData  `json:"bagData"`
	MailData AllMailData `json:"mailData"`
}

type cultivationData struct {
}

type activityData struct {
}

func (dao *PlayerDao) Unmarshal() {
	utils.UnCompressAndUnmarshal(dao.GeneralData, &dao.GeneralD)
	utils.UnCompressAndUnmarshal(dao.CultivationData, &dao.CultivationD)
	utils.UnCompressAndUnmarshal(dao.ActivityData, &dao.ActivityD)
}

func (dao *PlayerDao) TableName() string {
	return "player_dao"
}

func (dao *PlayerDao) GetServerId() int32 {
	return dao.ServerId
}

func (this *PlayerDao) TomSimplePlayer() *Player {
	p := &Player{
		PlayerId:      this.PlayerId,
		UserId:        this.UserId,
		ServerId:      this.ServerId,
		Nickname:      this.Nickname,
		IsOnline:      this.IsOnline,
		OfflineAt:     utils.FormatTimestamp(this.OfflineAt),
		OnlineAt:      utils.FormatTimestamp(this.OnlineAt),
		LastOfflineAt: utils.FormatTimestamp(this.LastOfflineAt),
	}
	this.Unmarshal()
	return p
}

type SaveDaoD struct {
	SaveType int
	SaveKey  string
	SaveData *PlayerDao
}
