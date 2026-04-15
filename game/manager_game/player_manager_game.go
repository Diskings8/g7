package manager_game

import (
	"g7/common/cronx"
	"g7/common/protos/pb"
	"g7/game/global_game"
	"g7/game/model_game"
	"sync/atomic"
)

var GPlayerManager = &playerManager{}

func NewPlayerBase(p *model_game.Player, StreamConn pb.GameStreamService_StreamServer, cancelFunc func()) {
	onlineData := model_game.OnlineData{}
	onlineData.Init(StreamConn, GSaveSystemManager.AsyncSaveQueue)
	p.OnlineData = onlineData
	p.StreamCancelFunc = cancelFunc
	p.CurRedisLockKey = global_game.MakePlayerRedisLockKey(p.ServerId, p.PlayerId)
	return
}

func OnLineRunning(p *model_game.Player) {
	//关键：启动玩家专属协程
	go p.RunMainRoutine()
}

type playerManager struct {
	cacheLockKeyReNewerSlot int32
	heartbeatSlot           int32
	cacheWriteBackSlot      int32
	dbWriteSlot             int32
}

func (p *playerManager) Init() {
	// 玩家锁续约
	cronx.AddPer1SecondTask(func() {
		slot := atomic.AddInt32(&p.cacheLockKeyReNewerSlot, 1) - 1
		curSlot := slot % 5
		global_game.GPlayerMaps.StartLockReNewer(curSlot)
	})
	// 心跳定时器
	cronx.AddPer12SecondTask(func() {
		slot := atomic.AddInt32(&p.heartbeatSlot, 1) - 1
		curSlot := slot % 5
		global_game.GPlayerMaps.HeartBeatCheck(curSlot)
	})
	// Redis回写定时器
	cronx.AddPer10SecondTask(func() {
		slot := atomic.AddInt32(&p.cacheWriteBackSlot, 1) - 1
		curSlot := slot % 5
		global_game.GPlayerMaps.RedisReWriteCheck(curSlot)
	})
	// 写入数据库
	cronx.AddPer30SecondTask(func() {
		slot := atomic.AddInt32(&p.dbWriteSlot, 1) - 1
		curSlot := slot % 5
		global_game.GPlayerMaps.DbWriteCheck(curSlot)
	})
}
