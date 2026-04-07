package redisx

import "strconv"

const (
	PlayerCacheKeyPrefix = "player:data:"
)

func MakePlayerCacheKey(playerId int64) string {
	return PlayerCacheKeyPrefix + strconv.FormatInt(playerId, 10)
}
