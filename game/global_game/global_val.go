package global_game

import (
	"fmt"
	"g7/common/dbc/dbc_interface"
	"g7/common/globals"
	"g7/common/mqc/mqc_interface"
	"g7/common/redisx"
	"g7/game/model_game"
)

var GGameDB dbc_interface.DBInterface
var GGlobalDB dbc_interface.DBInterface

func AutoMigrate(dbc dbc_interface.DBInterface) {
	if globals.IsDev() {
		_ = dbc.AutoMigrate(&model_game.PlayerDao{})
	}
}

var GGlobalMQ mqc_interface.MQProducerInterface

func MakePlayerRedisLockKey(serverId int32, playerId int64) string {
	return fmt.Sprintf(redisx.PlayerLockKeyPrefix, serverId, playerId)
}
