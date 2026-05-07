package room_server

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/rooms"
)

var GRoomMixServer = &RoomMixServer{}

type RoomMixServer struct {
	pb.UnimplementedRoomNodeServiceServer
	pb.UnimplementedRoomStreamServiceServer
	etcdAddr string
	roomMaps map[string]*rooms.Room
}
