package global_game

import (
	"g7/common/utils"
	"g7/game/model_game"
	"sync"
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
