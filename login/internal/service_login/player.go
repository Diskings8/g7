package service_login

import (
	"errors"
	"g7/common/snowflake"
	"g7/login/internal/dao_login"
	"g7/login/model_login"
)

// CreatePlayer 创建角色
func CreatePlayer(userID int64, nickname string) (*model_login.Player, error) {
	// 生成雪花UID
	uid := snowflake.GenUID()

	player := &model_login.Player{
		UID:      uid,
		UserID:   userID,
		Nickname: nickname,
	}

	if err := dao_login.CreatePlayer(player); err != nil {
		return nil, err
	}

	return player, nil
}

// SelectPlayer 选角（校验权限）
func SelectPlayer(userID int64, uid int64) (*model_login.Player, error) {
	player, err := dao_login.GetPlayerByUID(uid)
	if err != nil {
		return nil, errors.New("角色不存在")
	}

	// 校验角色属于该账号
	if player.UserID != userID {
		return nil, errors.New("无权限访问该角色")
	}

	return player, nil
}

func ListPlayersByUserID(userID int64) ([]*model_login.Player, error) {
	playerList, err := dao_login.ListPlayersByUserID(userID)
	return playerList, err
}
