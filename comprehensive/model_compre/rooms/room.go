package rooms

import (
	"context"
	"g7/common/maps"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle/actors"
	"g7/comprehensive/model_compre/battle/world"
	"github.com/golang/protobuf/proto"
	"sync"
	"time"
)

type PlayAction struct {
	PlayerId int64
	Action   *pb.GameMessage
}

type Room struct {
	mu            sync.RWMutex
	RoomId        string
	ConfId        int32
	players       map[int64]*RoomPlayerData
	members       []string
	inputChan     chan PlayAction
	tickRate      time.Duration // e.g. 50 * time.Millisecond
	pendingInputs []PlayAction
	world         *world.World
}

func NewRoom(confId int32, roomId string, members []string) *Room {
	r := &Room{
		ConfId:        confId,
		RoomId:        roomId,
		players:       make(map[int64]*RoomPlayerData),
		members:       members,
		inputChan:     make(chan PlayAction, 2000),
		pendingInputs: make([]PlayAction, 0, 2000),
		tickRate:      50 * time.Millisecond,
	}
	mapD := maps.GTiledManager.GetMap("FlowerRoom")
	r.world = world.NewWorld(r, mapD)
	return r
}

func (r *Room) GetPlayerStream(playerId int64) (*PlayerStream, bool) {
	val, ok := r.players[playerId]
	return val.Ps, ok
}

func (r *Room) SetPlayerStream(playerId int64, stream pb.RoomStreamService_StreamServer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rpd := NewRoomPlayerData(playerId, stream, r)
	r.players[playerId] = rpd
	go rpd.Ps.RunSendRoutine()
	return
}

func (r *Room) RemovePlayer(playerId int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, playerId)
	r.world.RemoveActor(playerId)
	if len(r.players) == 0 {
		r.close() // 触发房间主循环退出
	}
}

func (r *Room) EnterPlayerActor(playerId int64, actorInfo *pb.BattleActor) {
	worldActor := actors.NewPlayerActor(playerId, actorInfo, r.world)
	r.world.AddActor(&worldActor)
	r.world.InitActorGrid(&worldActor)
}

func (r *Room) Start(ctx context.Context) {
	ticker := time.NewTicker(r.tickRate) // 50ms
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-r.inputChan:
			if !ok {
				return
			}
			r.world.PushInput(msg.PlayerId, msg.Action)
		case <-ticker.C:
			r.update() //
		case <-ctx.Done():
			// 房间被主动关闭，做好清理
			r.shutdown()
			return
		}
	}
}

func (r *Room) shutdown() {
	r.mu.Lock()
	for _, rpd := range r.players {
		rpd.Ps.Quit()
	}
	r.mu.Unlock()
	close(r.inputChan)
}

func (r *Room) Broadcast(msg *pb.GameMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rpd := range r.players {
		select {
		case rpd.Ps.send <- msg:
		default: // 发送失败或通道满，跳过
		}
	}
}

func (r *Room) close() {

}

func (r *Room) shouldSend(playerId int64, msg *pb.GameMessage) bool {
	// todo 后续可以添加msg.data的hash码对比
	return true
}

func (r *Room) marshalGameMessage(MsgId pb.MsgID, data proto.Message) *pb.GameMessage {
	d, _ := proto.Marshal(data)
	return &pb.GameMessage{MsgId: uint32(MsgId), Body: d}
}

func (r *Room) calcNineGirdsView() {
	r.mu.RLock()
	players := make(map[int64]*RoomPlayerData, len(r.players))
	for k, v := range r.players {
		players[k] = v
	}
	r.mu.RUnlock()
	for playerID, rpd := range players {
		currentViewActors, currentViewEvents := r.world.GetNineGridsViewObjects(playerID)
		var enterList []int64  // 新进入
		var leaveList []int64  // 离开（可以不发，或发个 Leave 事件）
		var updateList []int64 // 留在视野内且变脏的
		lastView := rpd.GetLastView()

		enterList, updateList, leaveList = r.world.GetDirtyActorsInView(currentViewActors, lastView)
		// 无变化直接跳过
		if len(enterList) == 0 && len(updateList) == 0 && len(leaveList) == 0 && len(currentViewEvents) == 0 {
			continue
		}

		snap := r.world.NewWorldSnapshot(enterList, updateList, leaveList, currentViewEvents)
		msg := r.marshalGameMessage(pb.MsgID_MSG_World_NineGridsSnapshot, snap)
		select {
		case rpd.Ps.send <- msg:
		default:
		}
		rpd.SetLastView(currentViewActors)
	}
}

func (r *Room) update() {
	// 推动下一跳
	r.world.Tick(r.tickRate)

	defer func() {
		// 清除本帧标记
		r.world.ClearDirtyFlags()
	}()

	//各个角色的9宫格广播
	r.calcNineGirdsView()
}
