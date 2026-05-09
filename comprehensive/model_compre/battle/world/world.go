package world

import (
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/grids"
	"g7/comprehensive/model_compre/battle/interfaces"
	"github.com/golang/protobuf/proto"
	"sync"
	"sync/atomic"
)

type World struct {
	room      interfaces.Room
	running   atomic.Bool
	gridMap   grids.GirdMap
	actors    map[int64]interfaces.Actor    // 所有战斗实体（玩家、怪物等）
	actionsCh chan actoractions.ActorAction // 带缓冲
	eventLog  []Event                       // 本帧产生的逻辑事件（用于下发）
	frameID   uint32
	mu        sync.RWMutex
}

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

func (w *World) PushInput(actorId int64, action *pb.GameMessage) {
	select {
	case w.actionsCh <- actoractions.ActorAction{ActorId: actorId, Action: action}:
	default:
		logger.Log.Info("PushInput default")
		// 缓冲区满可丢弃或记录
	}
}

func (w *World) GetSnapshot() (rsp *pb.GameMessage, ok bool) {
	rsp = &pb.GameMessage{}
	ok = true
	snapshot := &pb.WorldSnapshot{
		FrameId:     w.frameID,
		ActorStates: []*pb.ActorState{},
	}
	w.mu.RLock()
	for _, actor := range w.actors {
		if !actor.IsDirty() {
			continue
		}
		protoActor := actor.ToProto()
		snapshot.ActorStates = append(snapshot.ActorStates, protoActor)
	}
	w.mu.RUnlock()
	if len(snapshot.ActorStates) == 0 {
		ok = false
	}
	rsp.MsgId = uint32(pb.MsgID_MSG_World_NineGridsSnapshot)
	rsp.Body, _ = proto.Marshal(snapshot)
	return
}

func (w *World) GetViewSnapshot(playerId int64) (rsp *pb.GameMessage) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	actor, ok := w.actors[playerId]

	if !ok {
		return nil
	}
	gx, gy := w.gridMap.GridCoord(actor.Pos())
	nGrids := w.gridMap.NineGrids(gx, gy)
	inView := make(map[int64]struct{})
	for _, g := range nGrids {
		for _, id := range w.gridMap.GetKeyActor(g) {
			inView[id] = struct{}{}
		}
	}
	snap := &pb.WorldSnapshot{FrameId: w.frameID}
	for id := range inView {
		if a, ok := w.actors[id]; ok {
			snap.ActorStates = append(snap.ActorStates, a.ToProto())
		}
	}
	rsp = &pb.GameMessage{MsgId: uint32(pb.MsgID_MSG_World_NineGridsSnapshot)}
	rsp.Body, _ = proto.Marshal(snap)
	return rsp
}
