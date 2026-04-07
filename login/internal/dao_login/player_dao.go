package dao_login

import (
	"g7/common/model_common"
	"g7/login/global_login"
)

func ListPlayersByUserID(userID int64) ([]*model_common.GlobalPlayerIndex, error) {
	var list []*model_common.GlobalPlayerIndex
	err := global_login.GLoginDB.FindList(&list, map[string]any{"user_id": userID})
	return list, err
}

func GetPlayerByUID(playerId int64) (*model_common.GlobalPlayerIndex, error) {
	var player model_common.GlobalPlayerIndex
	err := global_login.GLoginDB.FindOne(&player, map[string]any{"player_id": playerId})
	return &player, err
}
