package global_game

import (
	"encoding/json"
	"fmt"
	"g7/common/redisx"
	"g7/game/model_game"
	"time"
)

// 全局单例
var GPlayerCache = &PlayerCache{}

// Redis 缓存前缀
const (
	cacheExpire = 24 * time.Hour // 24小时过期（可调整）
)

// PlayerCache 玩家热数据缓存
type PlayerCache struct {
}

func (pc *PlayerCache) Init() {
}

// GetPlayerCache
// 1. 从 Redis 取
// 2. 反序列化为 PlayerDao
// 3. 不存在/失败 → 返回 error
func (pc *PlayerCache) GetPlayerCache(playerId int64) (*model_game.PlayerDao, error) {
	key := redisx.MakePlayerCacheKey(playerId)

	// 从 Redis 读取
	strData, err := redisx.GetKey(key)
	if err != nil {
		return nil, err // 不存在/失败 → 上层自动走 DB
	}

	// 反序列化为 DAO
	var dao model_game.PlayerDao
	if err := json.Unmarshal([]byte(strData), &dao); err != nil {
		return nil, err
	}

	return &dao, nil
}

// SetPlayerCache
// 保存玩家热数据到 Redis
func (pc *PlayerCache) SetPlayerCache(dao *model_game.PlayerDao) error {
	if dao == nil || dao.PlayerId == 0 {
		return fmt.Errorf("dao is nil or playerId=0")
	}

	// 序列化为 JSON
	data, err := json.Marshal(dao)
	if err != nil {
		return err
	}

	key := redisx.MakePlayerCacheKey(dao.PlayerId)
	// 保存 Redis + 设置过期
	return redisx.SetKey(key, data, cacheExpire)
}
