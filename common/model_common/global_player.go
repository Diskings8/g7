package model_common

type GlobalPlayerIndex struct {
	ID       int64 `gorm:"primaryKey"`
	UserID   int64 `gorm:"index" column:"user_id"`
	PlayerId int64 `column:"player_id"`
	ServerID int32
	Nickname string
}

func (GlobalPlayerIndex) TableName() string {
	return "global_player_index"
}
