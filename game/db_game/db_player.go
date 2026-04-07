package db_game

import (
	"g7/common/logger"
	"g7/game/global_game"
	"g7/game/model_game"
)

func GetPlayerByID(playerId int64) (*model_game.PlayerDao, error) {
	val := &model_game.PlayerDao{}
	err := global_game.GGameDB.FindOne(val, map[string]any{"player_id": playerId})
	return val, err
}

func SetPlayerDao(dao *model_game.PlayerDao) {
	err := global_game.GGameDB.Insert(dao)
	if err != nil {
		logger.Log.Warn(err.Error())
	}
}
