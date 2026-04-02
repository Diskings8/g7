package model

import "time"

type Player struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Uid       int64     `gorm:"column:uid;unique"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Nickname  string    `gorm:"column:nickname"`
	Level     int       `gorm:"column:level"`
	Exp       int64     `gorm:"column:exp"`
}

func (Player) TableName() string {
	return "player"
}
