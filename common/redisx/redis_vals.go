package redisx

import (
	"fmt"
	"strings"
)

func MakePlayerLoginValue(serverId int32, playerId int64, gameAddr string) string {
	return fmt.Sprintf("%d#%d#%s", serverId, playerId, gameAddr)
}

const (
	RedisPlayerLoginValIndexServer         = 0
	RedisPlayerLoginValIndexPlayerId       = 1
	RedisPlayerLoginValIndexGameServerAddr = 2
)

func GetPlayerLoginValueIndex(loginVal string, index int) string {
	return strings.Split(loginVal, "#")[index]
}

func MakeRoomMasterValue(roomId, roomAddr string) string {
	return fmt.Sprintf("%s#%s", roomId, roomAddr)
}

const (
	RedisRoomMasterValueIndexRoomId = 0
	RedisRoomMasterValueIndexAddr   = 1
)

func GetRoomMasterValueIndex(roomValue string, index int) string {
	return strings.Split(roomValue, "#")[index]
}
