package redisx

import (
	"fmt"
	"strings"
)

func MakePlayerLoginValue(serverId int32, playerId int64, gameAddr string) string {
	return fmt.Sprintf("%d_%d_%s", serverId, playerId, gameAddr)
}

const (
	RedisPlayerLoginValIndexServer         = 0
	RedisPlayerLoginValIndexPlayerId       = 1
	RedisPlayerLoginValIndexGameServerAddr = 2
)

func GetPlayerLoginValueIndex(loginVal string, index int) string {
	return strings.Split(loginVal, "_")[index]
}
