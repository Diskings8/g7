package world

import (
	"g7/common/logger"
	"g7/common/maps/maps_data"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/common_battle"
	"g7/comprehensive/model_compre/battle/events"
	"g7/comprehensive/model_compre/battle/grids"
	"g7/comprehensive/model_compre/battle/interfaces"
	"sync"
	"sync/atomic"
)

type World struct {
	room      interfaces.Room
	running   atomic.Bool
	gridMap   grids.GridMap
	mapD      maps_data.MapData
	actors    map[int64]interfaces.Actor    // 所有战斗实体（玩家、怪物等）
	actionsCh chan actoractions.ActorAction // 带缓冲
	frameID   int64
	mu        sync.RWMutex
}

var _ interfaces.World = (*World)(nil)

func NewWorld(room interfaces.Room, mapD maps_data.MapData) *World {
	w := &World{
		actors:    make(map[int64]interfaces.Actor),
		actionsCh: make(chan actoractions.ActorAction, 2000),
		mapD:      mapD,
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

func (w *World) NewWorldSnapshot(enterList, updateList, leaveList []int64, eventList []events.Event) (result *pb.WorldSnapshot) {
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
	for _, e := range eventList {
		result.Events = append(result.Events, e.ToProto())
	}
	return
}

func (w *World) PushInput(actorId int64, action *pb.GameMessage) {
	//logger.Log.Info(fmt.Sprintf("recv player:%v", action))
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

func (w *World) ClearTickGird() {

}

func (w *World) GetNineGridsViewObjects(playerId int64) (viewActors map[int64]struct{}, viewEvents []events.Event) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	actor, ok := w.actors[playerId]

	if !ok {
		return nil, nil
	}
	gx, gy := w.gridMap.GridCoord(actor.GetPos().CurPos)
	nGrids := w.gridMap.NineGrids(gx, gy)

	viewActors = make(map[int64]struct{})
	viewEvents = make([]events.Event, 0)

	for _, g := range nGrids {
		for _, id := range w.gridMap.GetGridActors(g[0], g[1]) {
			viewActors[id] = struct{}{}
		}
		girdEvents := w.gridMap.GetGridEvent(g[0], g[1])
		viewEvents = append(viewEvents, girdEvents...)
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
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, v := range w.actors {
		targetActors = append(targetActors, v)
		return targetActors
	}
	return nil
}

func (w *World) AddEvent(pos common_battle.Vector3D, event events.Event) {
	x, y := w.gridMap.GridCoord(pos)
	w.gridMap.SetGridEvent(x, y, event)
}

func (w *World) CheckPosBlock(pos common_battle.Vector3D) bool {
	return maps_data.IsBlock(w.mapD, pos.X, pos.Y)
}
