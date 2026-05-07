package rooms

import (
	"g7/common/protos/pb"
	"sync"
)

type PlayerStream struct {
	playerId int64
	send     chan *pb.GameMessage
	stop     chan struct{}
	stream   pb.RoomStreamService_StreamServer
	room     *Room
	quitOnce sync.Once
}

func NewPlayerStream(playerId int64, stream pb.RoomStreamService_StreamServer, room *Room) *PlayerStream {
	return &PlayerStream{playerId: playerId, stream: stream,
		send: make(chan *pb.GameMessage, 300),
		stop: make(chan struct{}),
		room: room}
}

func (ps *PlayerStream) Recv(msg *pb.GameMessage) {
	select {
	case ps.room.inputChan <- PlayAction{PlayerId: ps.playerId, Action: msg}:
	case <-ps.stop:
		// 玩家已退出，静默丢弃
	}
}

func (ps *PlayerStream) Quit() {
	ps.room.RemovePlayer(ps.playerId)
	ps.quitOnce.Do(func() {
		close(ps.stop) // 广播退出信号
	})
}

func (ps *PlayerStream) RunSendRoutine() {
	defer func() {
		close(ps.send)
	}()
	for {
		select {
		case msg := <-ps.send:
			ps.stream.Send(msg)
		case <-ps.stop:
			return
		}
	}
}
