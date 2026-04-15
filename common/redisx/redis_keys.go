package redisx

import (
	"fmt"
)

const (
	PlayerCacheKeyPrefix = "player:data:%d:%d"
	PlayerLockKeyPrefix  = "player:lock:%d:%d"
	ShopLockKeyPrefix    = "shop:lock:%d:%d"
)

func MakePlayerCacheKey(serverId int32, playerId int64) string {
	return fmt.Sprintf(PlayerCacheKeyPrefix, serverId, playerId)
}
