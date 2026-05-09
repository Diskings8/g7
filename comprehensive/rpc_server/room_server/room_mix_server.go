package room_server

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/rooms"
	"sync"
)

var GRoomMixServer = &RoomMixServer{}

type RoomMixServer struct {
	pb.UnimplementedRoomNodeServiceServer
	pb.UnimplementedRoomStreamServiceServer
	etcdAddr string
	roomMaps map[string]*rooms.Room
	rw       sync.RWMutex
}

func (rms *RoomMixServer) AddRoom(room *rooms.Room) {
	rms.rw.Lock()
	rms.roomMaps[room.RoomId] = room
	rms.rw.Unlock()
}

func (rms *RoomMixServer) GetRoom(roomId string) (r *rooms.Room, ok bool) {
	rms.rw.RLock()
	defer rms.rw.RUnlock()
	r, ok = rms.roomMaps[roomId]
	return
}
