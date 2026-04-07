package global_game

import (
	"g7/common/dbc"
	"g7/common/utils"
	"g7/game/model_game"
)

var GGameDB dbc.DBInterface
var GGlobalDB dbc.DBInterface

func AutoMigrate(dbc dbc.DBInterface) {
	if utils.IsDev() {
		_ = dbc.AutoMigrate(&model_game.PlayerDao{})
	}
}
