package rooms

import (
	"context"
	"fmt"
	"g7/common/logger"
	"g7/common/protos/pb"
	"sync"
	"time"
)

type PlayAction struct {
	PlayerId int64
	Action   *pb.GameMessage
}

type Room struct {
	mu            sync.RWMutex
	roomId        string
	players       map[int64]*PlayerStream
	members       []string
	inputChan     chan PlayAction
	tickRate      time.Duration // e.g. 50 * time.Millisecond
	pendingInputs []PlayAction
	world         World
}

func NewRoom(confId int32, roomId string, members []string) *Room {
	return &Room{roomId: roomId, players: make(map[int64]*PlayerStream),
		members: members, inputChan: make(chan PlayAction, 2000),
		pendingInputs: make([]PlayAction, 0, 2000),
		tickRate:      50 * time.Millisecond}
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
	if len(r.players) == 0 {
		r.close() // 触发房间主循环退出
	}
}

func (r *Room) Start(ctx context.Context) {
	logger.Log.Info(fmt.Sprintf("%s room start", r.roomId))
	ticker := time.NewTicker(r.tickRate)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-r.inputChan:
			if !ok {
				return
			}
			r.pendingInputs = append(r.pendingInputs, msg)
		case <-ticker.C:
			r.update()
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

func (r *Room) update() {
	for _, input := range r.pendingInputs {
		r.handleInput(input)
	}
	r.pendingInputs = r.pendingInputs[:0]

	// 2. 推进世界状态（移动、技能冷却、AI、计时器等）
	r.world.Step(r.tickRate)

	// 3. 生成下行广播数据
	snapshot := r.world.GetSnapshot()

	// 4. 广播给所有玩家
	r.broadcast(snapshot)
}

func (r *Room) broadcast(msg *pb.GameMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, session := range r.players {
		select {
		case session.send <- msg:
		default: // 发送失败或通道满，跳过
		}
	}
}

func (r *Room) handleInput(input PlayAction) {

}

func (r *Room) close() {

}
