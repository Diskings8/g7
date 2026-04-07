package model_game

import "time"

type ItemData struct {
	CfgID      int32     `gorm:"column:cfg_id;index" bson:"cfg_id"`       // 配置表ID
	IsBind     int32     `gorm:"column:is_bind" bson:"is_bind"`           // 0=非绑定 1=绑定
	IsUnique   int32     `gorm:"column:is_unique" bson:"is_unique"`       // 0=非唯一 1=唯一
	ExpireType int32     `gorm:"column:expire_type" bson:"expire_type"`   // 过期类型 0=永久 1=限时
	Num        int64     `gorm:"column:num" bson:"num"`                   // 数量（堆叠用）
	OwnerID    int64     `gorm:"column:owner_id;index" bson:"owner_id"`   // 归属者id
	CreateID   int64     `gorm:"column:create_id;index" bson:"create_id"` // 创造者id
	UniqueID   uint64    `gorm:"column:unique_id" bson:"unique_id"`       // 唯一ID
	ExpireTime time.Time `gorm:"column:expire_time" bson:"expire_time"`   // 过期时间
	CreateTime time.Time `gorm:"column:create_time" bson:"create_time"`   // 创建时间
}
