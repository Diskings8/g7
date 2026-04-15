package global_game

import (
	"context"
	"errors"
	"g7/common/configx"
	"g7/common/globals"
	"g7/game/model_game"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
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
	this.lockExpire = 10 * time.Second
	this.renewInterval = 10 * time.Second
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

func (this *playerMaps) SetPlayer(playerId int64, h *model_game.Player) {
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

func (this *playerMaps) StartLockReNewer(curSlot int32) {
	this.rwLock.RLock()
	capLen := len(this.Data) / 5
	if capLen == 0 {
		capLen = 16
	} // 兜底初始容量
	players := make([]*model_game.Player, 0, capLen)
	//logger.Log.Info(fmt.Sprintf("%d", curSlot))
	for _, v := range this.Data {
		if int32(v.PlayerId%5) != curSlot {
			continue
		}
		players = append(players, v)
	}
	this.rwLock.RUnlock()

	playerLoseLockL := make([]*model_game.Player, 0, 10)
	for _, v := range players {
		if ok := this.CheckLockValid(v.ServerId, v.PlayerId); ok {
			_ = this.renewLock(v.ServerId, v.PlayerId)
			//logger.Log.Info(fmt.Sprintf("renewLock: %d,%d,%v", v.PlayerId, curSlot, b))
		} else {
			//logger.Log.Warn(fmt.Sprintf("%d,CheckLockBelongsToMe", v.PlayerId))
			// 锁不属于我 → 玩家已被顶号/转移 → 主动下线
			playerLoseLockL = append(playerLoseLockL)
		}
	}
	// ===========================
	this.DelPlayerList(playerLoseLockL)
}

func (this *playerMaps) CheckLockBelongsToMe(serverId int32, playerId int64) (bool, error) {
	key := MakePlayerRedisLockKey(serverId, playerId)
	// SETNX 原子加锁
	return this.cli.SetNX(context.Background(), key, globals.InstanceId, this.lockExpire).Result()
}

// renewLock 续约锁
func (this *playerMaps) renewLock(serverId int32, playerId int64) bool {
	key := MakePlayerRedisLockKey(serverId, playerId)
	// 先校验锁归属，再续约
	val, err := this.cli.Get(context.Background(), key).Result()
	if err != nil || val != globals.InstanceId {
		return false
	}
	b := this.cli.Expire(context.Background(), key, this.renewInterval)
	return b.Val()
}

// 收包前必校验锁（核心！）
func (this *playerMaps) CheckLockValid(serverId int32, playerId int64) bool {
	key := MakePlayerRedisLockKey(serverId, playerId)
	val, err := this.cli.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return false // 锁不存在
	}
	if err != nil {
		return false
	}
	return val == globals.InstanceId
}
