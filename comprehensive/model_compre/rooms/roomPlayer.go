package rooms

import "g7/common/protos/pb"

type RoomPlayerData struct {
	Ps       *PlayerStream
	LastView map[int64]struct{}
}

func NewRoomPlayerData(playerId int64, stream pb.RoomStreamService_StreamServer, room *Room) *RoomPlayerData {
	rpd := &RoomPlayerData{}
	rpd.Ps = NewPlayerStream(playerId, stream, room)
	rpd.LastView = make(map[int64]struct{})
	return rpd
}

func (rpd *RoomPlayerData) GetLastView() map[int64]struct{} {
	return rpd.LastView
}

func (rpd *RoomPlayerData) SetLastView(newView map[int64]struct{}) {
	if newView == nil {
		clear(rpd.LastView)
		return
	}
	rpd.LastView = newView
}
