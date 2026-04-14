package service_login

import (
	"errors"
	"g7/common/model_common"
	"g7/login/internal/dao_login"
)

// SelectPlayer 选角（校验权限）
func SelectPlayer(userID int64, playerID int64) (*model_common.GlobalPlayerIndex, error) {
	player, err := dao_login.GetPlayerByUID(playerID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}

	// 校验角色属于该账号
	if player.UserID != userID {
		return nil, errors.New("无权限访问该角色")
	}

	return player, nil
}

func ListPlayersByUserID(userID int64) ([]*model_common.GlobalPlayerIndex, error) {
	playerList, err := dao_login.ListPlayersByUserID(userID)
	return playerList, err
}
