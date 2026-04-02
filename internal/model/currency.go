package model

import "time"

// Currency 货币类型
type Currency struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Uid       int64     `gorm:"column:uid;uniqueIndex"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Gold      int64     `gorm:"column:gold"`
	Diamond   int64     `gorm:"column:diamond"`
}

func (Currency) TableName() string {
	return "currency"
}
