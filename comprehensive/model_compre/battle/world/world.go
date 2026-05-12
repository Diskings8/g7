package world

import (
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/events"
	"g7/comprehensive/model_compre/battle/grids"
	"g7/comprehensive/model_compre/battle/interfaces"
	"sync"
	"sync/atomic"
)

type World struct {
	room      interfaces.Room
	running   atomic.Bool
	gridMap   grids.GirdMap
	actors    map[int64]interfaces.Actor    // 所有战斗实体（玩家、怪物等）
	actionsCh chan actoractions.ActorAction // 带缓冲
	eventLog  []events.Event                // 本帧产生的逻辑事件（用于下发）
	frameID   int64
	mu        sync.RWMutex
}

var _ interfaces.World = (*World)(nil)

func NewWorld(room interfaces.Room) *World {
	w := &World{
		actors:    make(map[int64]interfaces.Actor),
		actionsCh: make(chan actoractions.ActorAction, 2000),
		room:      room,
	}
	w.gridMap.Init()
	return w
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

func (w *World) GetDirtyActorsInView(currentView map[int64]struct{}, lastView map[int64]struct{}) (enter, update, leave []int64) {
	if currentView == nil {
		return nil, nil, nil
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	for id := range currentView {
		if _, ok := lastView[id]; ok {
			if x := w.actors[id]; x != nil && x.IsDirty() {
				update = append(update, id)
			}
		} else {
			enter = append(enter, id) // 新进入
		}
	}
	for id := range lastView {
		if _, ok := currentView[id]; !ok {
			leave = append(leave, id)
		}
	}
	return
}

func (w *World) NewWorldSnapshot(enterList, updateList, leaveList []int64) (result *pb.WorldSnapshot) {
	result = &pb.WorldSnapshot{FrameId: w.frameID, LeaveList: leaveList}
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, oneActorId := range enterList {
		ac, ok := w.actors[oneActorId]
		if !ok {
			continue
		}
		result.EnterList = append(result.EnterList, ac.ToProto(true))
	}
	for _, oneActorId := range updateList {
		ac, ok := w.actors[oneActorId]
		if !ok {
			continue
		}
		result.UpdateList = append(result.UpdateList, ac.ToProto(false))
	}
	return
}

func (w *World) PushInput(actorId int64, action *pb.GameMessage) {
	select {
	case w.actionsCh <- actoractions.ActorAction{ActorId: actorId, Action: action}:
	default:
		logger.Log.Info("PushInput default")
		// 缓冲区满可丢弃或记录
	}
}

func (w *World) ClearDirtyFlags() {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, actor := range w.actors {
		actor.ClearDirty()
	}
}

func (w *World) GetNineGridsViewActors(playerId int64) (viewActors map[int64]struct{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	actor, ok := w.actors[playerId]

	if !ok {
		return nil
	}
	gx, gy := w.gridMap.GridCoord(actor.Pos())
	nGrids := w.gridMap.NineGrids(gx, gy)
	viewActors = make(map[int64]struct{})
	for _, g := range nGrids {
		for _, id := range w.gridMap.GetKeyActor(g) {
			viewActors[id] = struct{}{}
		}
	}
	return
}

func (w *World) getActor(actorId int64) interfaces.Actor {
	w.mu.RLock()
	defer w.mu.RUnlock()

	actor, ok := w.actors[actorId]
	if ok {
		return actor
	}
	return nil
}

func (w *World) distance(a, b interfaces.Actor) float64 {
	//a.Pos(),b.Pos()
	return 0
}

func (w *World) FindActors(src interfaces.Actor, actorIds []int64, params ...any) []interfaces.Actor {
	var targetActors []interfaces.Actor
	if actorIds != nil {
		for _, v := range actorIds {
			t := w.getActor(v)
			if t == nil {
				continue
			}
			if w.distance(t, src) <= params[0].(float64) {
				targetActors = append(targetActors, t)
			}
		}
		return targetActors
	}
	return nil
}

func (w *World) AddEvent(event events.Event) {
	w.eventLog = append(w.eventLog, event)
}
