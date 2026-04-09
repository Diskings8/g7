package global_game

import (
	"g7/common/dbc/dbc_interface"
	"g7/common/mqc/mqc_interface"
	"g7/common/utils"
	"g7/game/model_game"
)

var GGameDB dbc_interface.DBInterface
var GGlobalDB dbc_interface.DBInterface

func AutoMigrate(dbc dbc_interface.DBInterface) {
	if utils.IsDev() {
		_ = dbc.AutoMigrate(&model_game.PlayerDao{})
	}
}

var GGlobalMQ mqc_interface.MQProducerInterface
