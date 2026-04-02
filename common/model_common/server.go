package model_common

type Server struct {
	ID        int64  `gorm:"primaryKey;column:id"`
	ServerID  int    `gorm:"column:server_id;uniqueIndex"`
	Name      string `gorm:"column:name"`
	Addr      string `gorm:"column:addr"`
	DBName    string `gorm:"column:db_name"`
	Status    int    `gorm:"column:status;default:1"`
	Channel   int    `gorm:"column:channel"`
	GroupID   int    `gorm:"column:group_id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
}

func (Server) TableName() string {
	return "server_list"
}
