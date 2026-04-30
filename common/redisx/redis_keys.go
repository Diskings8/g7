package redisx

import (
	"fmt"
)

const (
	PlayerCacheKeyPrefix = "player:data:%d:%d"
	PlayerLockKeyPrefix  = "player:lock:%d:%d"
	PlayerLoginKeyPrefix = "player:login:%d:%d"
	ShopLockKeyPrefix    = "shop:lock:%d:%d"
)

func MakePlayerCacheKey(serverId int32, playerId int64) string {
	return fmt.Sprintf(PlayerCacheKeyPrefix, serverId, playerId)
}

func MakePlayerLockKey(serverId int32, playerId int64) string {
	return fmt.Sprintf(PlayerLockKeyPrefix, serverId, playerId)
}

func MakePlayerLoginKey(serverId int32, playerId int64) string {
	return fmt.Sprintf(PlayerLoginKeyPrefix, serverId, playerId)
}
