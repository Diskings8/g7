package model_common

type GlobalPlayerIndex struct {
	ID       int64 `gorm:"primaryKey"`
	UID      int64 `gorm:"uniqueIndex"`
	UserID   int64 `gorm:"index"`
	ServerID int
	Nickname string
}

func (GlobalPlayerIndex) TableName() string {
	return "global_player_index"
}
