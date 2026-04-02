package model

import "time"

type User struct {
	ID          int64     `gorm:"primaryKey;column:id"`
	Username    string    `gorm:"column:username;unique"`
	Password    string    `gorm:"column:password"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	IsBand      bool      `gorm:"column:is_band"`
	ChannelType int32     `gorm:"column:channel_type"`
	ChannelID   string    `gorm:"column:channel_id"`
	PlayerIDs   string    `gorm:"column:players_ids"`
}

func (User) TableName() string {
	return "user"
}
