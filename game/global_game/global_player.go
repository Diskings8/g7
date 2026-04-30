package global_game

import (
	"context"
	"g7/common/configx"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/redisx"
	"g7/game/model_game"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var GPlayerMaps playerMaps

type playerMaps struct {
	Data   map[int64]*model_game.Player // playerID => PlayerHandle
	rwLock sync.RWMutex
	cli    *redis.Client

	lockExpire     time.Duration
	renewInterval  time.Duration
	offlineLockExt time.Duration
}

func (this *playerMaps) Init(redisCli *redis.Client) {
	this.Data = make(map[int64]*model_game.Player)
	this.cli = redisCli
	if globals.IsDev() {
		this.lockExpire = 10 * time.Minute
	} else {
		this.lockExpire = 10 * time.Second
	}
	if globals.IsDev() {
		this.renewInterval = 10 * time.Minute
	} else {
		this.renewInterval = 10 * time.Second
	}
	this.offlineLockExt = 10 * time.Minute
}

func (this *playerMaps) GetPlayer(playerId int64) *model_game.Player {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	v, ok := this.Data[playerId]
	if ok {
		return v
	}
	return nil
}

func (this *playerMaps) GetAllPlayerIds() []int64 {
	var result = make([]int64, 0, len(this.Data))
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	for k := range this.Data {
		result = append(result, k)
	}
	return result
}

func (this *playerMaps) SetPlayer(playerId int64, h *model_game.Player) {
	if playerId <= 0 {
		logger.Log.Warn("has emptyPlayerId")
		return
	}
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.Data[playerId] = h
}

func (this *playerMaps) DelOnePlayerById(playerId int64) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	delete(this.Data, playerId)
}

func (this *playerMaps) DelPlayerList(playerList []*model_game.Player) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	for _, v := range playerList {
		key := v.PlayerId
		// 仅当玩家存在时删除，避免无效操作
		if _, exists := this.Data[key]; exists {
			delete(this.Data, key)
		}
	}
}

func (this *playerMaps) DelAll() {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	clear(this.Data)
}

// HeartBeatCheck 剔除超时心跳的玩家
func (this *playerMaps) HeartBeatCheck(curSlot int32) {
	curTime := time.Now()
	checkBeatTime := time.Duration(configx.GEnvCfg.Env.HeatBeatSeconds)
	if checkBeatTime <= 0 {
		checkBeatTime = 60
	}

	this.rwLock.RLock()
	capLen := len(this.Data) / 5
	if capLen == 0 {
		capLen = 16
	} // 兜底初始容量
	players := make([]*model_game.Player, 0, capLen)

	for _, v := range this.Data {
		if int32(v.PlayerId%5) != curSlot {
			continue
		}
		if v.GetLastHearBeatTime().Add(checkBeatTime * time.Second).Before(curTime) {
			players = append(players, v)
		}
	}
	this.rwLock.RUnlock()
	for _, v := range players {
		v.RunInActor(func() {
			v.Kick("heartbeat timeout")
		})
	}
	this.DelPlayerList(players)
	this.DelRedisLoginKey(players)
}

// RedisReWriteCheck Redis定时回写任务
func (this *playerMaps) RedisReWriteCheck(curSlot int32) {
	this.rwLock.RLock()
	capLen := len(this.Data) / 5
	if capLen == 0 {
		capLen = 16
	} // 兜底初始容量
	players := make([]*model_game.Player, 0, capLen)

	for _, v := range this.Data {
		if int32(v.PlayerId%5) != curSlot {
			continue
		}
		players = append(players, v)
	}
	this.rwLock.RUnlock()

	for _, v := range players {
		v.RunInActor(func() {
			v.RedisReWrite(globals.SaveDataKindCornCache)
		})
	}

}

// DbWriteCheck 定时存库任务
func (this *playerMaps) DbWriteCheck(curSlot int32) {
	this.rwLock.RLock()
	capLen := len(this.Data) / 5
	if capLen == 0 {
		capLen = 16
	} // 兜底初始容量
	players := make([]*model_game.Player, 0, capLen)

	for _, v := range this.Data {
		if int32(v.PlayerId%5) != curSlot {
			continue
		}
		players = append(players, v)
	}
	this.rwLock.RUnlock()

	for _, v := range players {
		v.RunInActor(func() {
			v.DbWrite(globals.SaveDataKindCornDb)
		})
	}

}

func (this *playerMaps) AllRunFunc(fc func(*model_game.Player)) {
	this.rwLock.RLock()
	players := make([]*model_game.Player, 0, len(this.Data))
	for _, v := range this.Data {
		players = append(players, v)
	}
	this.rwLock.RUnlock()
	batchSize := 100
	for i := 0; i < len(players); i += batchSize {
		end := i + batchSize
		if end > len(players) {
			end = len(players)
		}
		batch := players[i:end]

		go func(batch []*model_game.Player) {
			for _, v := range batch {
				v.RunInActor(func() {
					fc(v)
				})
			}
		}(batch)
		time.Sleep(10 * time.Millisecond)
	}
}

func (this *playerMaps) DelRedisLoginKey(ps []*model_game.Player) {
	var keys []string
	for _, p := range ps {
		key := redisx.MakePlayerLoginKey(p.ServerId, p.PlayerId)
		keys = append(keys, key)
	}
	this.cli.Del(context.Background(), keys...)
}

func (this *playerMaps) RegisterRedisLoginKey(p *model_game.Player) {
	key := redisx.MakePlayerLoginKey(p.ServerId, p.PlayerId)
	this.cli.Set(context.Background(), key, globals.GetServerInstance(), time.Hour*7*24)
}
