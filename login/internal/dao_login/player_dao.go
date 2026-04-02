package dao_login

import (
	"g7/common/mysqlx"
	"g7/login/model_login"
)

func CreatePlayer(player *model_login.Player) error {
	return mysqlx.GlobalDb.Create(player).Error
}

func ListPlayersByUserID(userID int64) ([]*model_login.Player, error) {
	var list []*model_login.Player
	err := mysqlx.GlobalDb.Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

func GetPlayerByUID(uid int64) (*model_login.Player, error) {
	var player model_login.Player
	err := mysqlx.GlobalDb.Where("uid = ?", uid).First(&player).Error
	return &player, err
}
