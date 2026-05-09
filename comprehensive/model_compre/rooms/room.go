package rooms

import (
	"context"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle/actors"
	"g7/comprehensive/model_compre/battle/world"
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
	players       map[int64]*PlayerStream
	members       []string
	inputChan     chan PlayAction
	tickRate      time.Duration // e.g. 50 * time.Millisecond
	pendingInputs []PlayAction
	world         *world.World
}

func NewRoom(confId int32, roomId string, members []string) *Room {
	r := &Room{
		RoomId:        roomId,
		players:       make(map[int64]*PlayerStream),
		members:       members,
		inputChan:     make(chan PlayAction, 2000),
		pendingInputs: make([]PlayAction, 0, 2000),
	}
	r.world = world.NewWorld(r)
	return r
}

func (r *Room) GetPlayerStream(playerId int64) (*PlayerStream, bool) {
	val, ok := r.players[playerId]
	return val, ok
}

func (r *Room) SetPlayerStream(playerId int64, stream pb.RoomStreamService_StreamServer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ps := NewPlayerStream(playerId, stream, r)
	r.players[playerId] = ps
	go ps.RunSendRoutine()
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
	worldActor := actors.NewPlayerActor(playerId, actorInfo)
	r.world.AddActor(&worldActor)
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
	for _, session := range r.players {
		session.Quit()
	}
	r.mu.Unlock()
	close(r.inputChan)
}

func (r *Room) Broadcast(msg *pb.GameMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, session := range r.players {
		select {
		case session.send <- msg:
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

func (r *Room) update() {
	// 推动下一跳
	r.world.Tick(r.tickRate)

	//各个角色的9宫格广播
	r.mu.RLock()
	defer r.mu.RUnlock()
	for playerID, session := range r.players {
		msg := r.world.GetViewSnapshot(playerID)
		if msg != nil && r.shouldSend(playerID, msg) {
			select {
			case session.send <- msg:
			default:
			}
		}
	}
}
