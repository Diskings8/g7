package model_common

import (
	"g7/common/structs"
	"time"
)

type DBTableInterface interface {
	TableName() string
}

type ActionLog struct {
	ID           int64                  `gorm:"column:id;primaryKey;autoIncrement"`
	PlayerID     int64                  `gorm:"column:player_id"`
	Action       string                 `gorm:"column:action"`
	Reason       string                 `gorm:"column:reason"`
	CostItem     []structs.KInt32VInt64 `gorm:"column:cost_item;serializer:json"`
	CostCurrency []structs.KInt32VInt64 `gorm:"column:cost_currency;serializer:json"`
	GainItem     []structs.KInt32VInt64 `gorm:"column:gain_item;serializer:json"`
	GainCurrency []structs.KInt32VInt64 `gorm:"column:gain_currency;serializer:json"`
	Ext          string                 `gorm:"column:ext"`
	CreateTime   time.Time              `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	SaveTime     time.Time              `gorm:"column:save_time;default:CURRENT_TIMESTAMP"`
}

func (ActionLog) TableName() string {
	return "action_log"
}
