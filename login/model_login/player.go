package model_login

import "time"

type Player struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	PlayerID  int64     `gorm:"column:player_id;unique"` // 雪花
	UserID    int64     `gorm:"column:user_id"`          // 关联账号ID（user.id）
	ServerID  int64     `gorm:"column:server_id"`        // 区服id
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Nickname  string    `gorm:"column:nickname"`
}

func (Player) TableName() string {
	return "player"
}
