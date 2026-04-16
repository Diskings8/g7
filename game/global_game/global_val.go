package global_game

import (
	"fmt"
	"g7/common/dbc"
	"g7/common/dbc/dbc_interface"
	"g7/common/model_common"
	"g7/common/mqc/mqc_interface"
	"g7/common/redisx"
	"g7/game/model_game"
)

var GGameDB dbc_interface.DBInterface
var GGlobalDB dbc_interface.DBInterface

func AutoMigrate(dbi dbc_interface.DBInterface) {
	_ = dbc.AutoMigrates(dbi, &model_game.PlayerDao{}, &model_common.PlayerMail{})
}

var GGlobalMQ mqc_interface.MQProducerInterface

func MakePlayerRedisLockKey(serverId int32, playerId int64) string {
	return fmt.Sprintf(redisx.PlayerLockKeyPrefix, serverId, playerId)
}
