package global_game

import (
	"g7/common/configx"
	"g7/common/utils"
	"g7/game/model_game"
	"sync"
	"time"
)

var GPlayerMaps playerMaps

type playerMaps struct {
	Data   map[string]*model_game.Player // playerID => PlayerHandle
	rwLock sync.RWMutex
}

func (this *playerMaps) Init() {
	this.Data = make(map[string]*model_game.Player)
}

func (this *playerMaps) GetPlayer(playerId int64) *model_game.Player {
	tar := utils.Int64ToString(playerId)
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	v, ok := this.Data[tar]
	if ok {
		return v
	}
	return nil
}

func (this *playerMaps) SetPlayer(playerId int64, h *model_game.Player) {
	tar := utils.Int64ToString(playerId)
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.Data[tar] = h
}

func (this *playerMaps) DelPlayer(playerId int64) {
	tar := utils.Int64ToString(playerId)
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	delete(this.Data, tar)
}

func (this *playerMaps) DelAll() {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	clear(this.Data)
}

func (this *playerMaps) HeartBeatCheck() {
	curTime := time.Now()
	checkBeatTime := time.Duration(configx.GEnvCfg.Env.HeatBeatSeconds)
	if checkBeatTime <= 0 {
		checkBeatTime = 30
	}
	this.rwLock.RLock()
	players := make([]*model_game.Player, 0, len(this.Data))
	for _, v := range this.Data {
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
	this.rwLock.Lock()
	for _, v := range players {
		key := utils.Int64ToString(v.PlayerId)
		// 仅当玩家存在时删除，避免无效操作
		if _, exists := this.Data[key]; exists {
			delete(this.Data, key)
		}
	}
	this.rwLock.Unlock()
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
