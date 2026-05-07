package room_server

import (
	"context"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/rooms"
)

func (rms *RoomMixServer) Init(etcdAddr string) {
	rms.etcdAddr = etcdAddr
	rms.roomMaps = make(map[string]*rooms.Room)
}

func (rms *RoomMixServer) CreateRoom(_ctx context.Context, req *pb.Req_Node_CreateRoom) (*pb.Rsp_Node_CreateRoom, error) {
	roomId := req.GetMatchId()
	rsp := &pb.Rsp_Node_CreateRoom{
		RoomId:   roomId,
		State:    1,
		RoomAddr: rms.etcdAddr,
	}
	room := rooms.NewRoom(req.GetConfId(), roomId, req.GetMatchMember())
	rms.roomMaps[roomId] = room
	go room.Start()
	return rsp, nil
}
