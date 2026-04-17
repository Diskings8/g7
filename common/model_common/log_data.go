package model_common

import (
	"g7/common/structs"
)

type DBTableInterface interface {
	TableName() string
}

type DBMqInterface interface {
	TableName() string
	GetServerId() int32
	GetEventType() int32
}

type BaseLog struct {
	ID         int64 `gorm:"column:id;primaryKey;autoIncrement"`
	ServerId   int32 `gorm:"column:server_id"`
	EventType  int32 `gorm:"column:event_type"`
	CreateTime int64 `gorm:"column:create_time;"`
	SaveTime   int64 `gorm:"column:save_time;" json:"-"`
}

func (l BaseLog) GetServerId() int32 {
	return l.ServerId
}
func (l BaseLog) GetEventType() int32 {
	return l.EventType
}

type ActionLog struct {
	BaseLog
	PlayerID     int64                      `gorm:"column:player_id"`
	Action       string                     `gorm:"column:action"`
	Reason       string                     `gorm:"column:reason"`
	CostItem     []structs.KInt32VInt64Bind `gorm:"column:cost_item;serializer:json"`
	CostCurrency []structs.KInt32VInt64Bind `gorm:"column:cost_currency;serializer:json"`
	GainItem     []structs.KInt32VInt64Bind `gorm:"column:gain_item;serializer:json"`
	GainCurrency []structs.KInt32VInt64Bind `gorm:"column:gain_currency;serializer:json"`
	Ext          string                     `gorm:"column:ext"`
}

func (ActionLog) TableName() string {
	return "action_log"
}
