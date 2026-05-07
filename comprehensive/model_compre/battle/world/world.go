package world

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/interfaces"
	"sync"
	"sync/atomic"
	"time"
)

type World struct {
	room      any //  *room.Room
	running   atomic.Bool
	actors    map[int64]interfaces.Actor    // 所有战斗实体（玩家、怪物等）
	actionsCh chan actoractions.ActorAction // 带缓冲
	eventLog  []Event                       // 本帧产生的逻辑事件（用于下发）
	frameID   uint32
	mu        sync.RWMutex
}

func NewWorld() *World {
	return &World{
		actors:    make(map[int64]interfaces.Actor),
		actionsCh: make(chan actoractions.ActorAction, 2000),
	}
}

func (w *World) Start() {
	w.running.Store(true)
	go w.runLoop()
}

// Stop 停止世界
func (w *World) Stop() {
	w.running.Store(false)
}

// AddActor 添加单位
func (w *World) AddActor(actor interfaces.Actor) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.actors[actor.ID()] = actor
}

// RemoveActor 移除单位
func (w *World) RemoveActor(actorID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.actors, actorID)
}

// runLoop 世界帧循环（真正跑逻辑的地方）
func (w *World) runLoop() {
	ticker := time.NewTicker(battle.FrameTime)
	defer ticker.Stop()

	for w.running.Load() {
		<-ticker.C
		w.Tick(battle.FrameTime)
	}
}

func (w *World) PushInput(actorId int64, action *pb.GameMessage) {
	select {
	case w.actionsCh <- actoractions.ActorAction{ActorId: actorId, Action: action}:
	default:
		// 缓冲区满可丢弃或记录
	}
}

func (w *World) GetSnapshot() *pb.GameMessage {
	return &pb.GameMessage{}
}
