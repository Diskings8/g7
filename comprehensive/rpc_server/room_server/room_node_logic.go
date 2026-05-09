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

func (rms *RoomMixServer) CreateRoom(_ctx context.Context, req *pb.Req_Node_CreateRoom) (rsp *pb.Rsp_Node_CreateRoom, e error) {
	roomId := req.GetMatchId()
	//logger.Log.Info(roomId)
	rsp = &pb.Rsp_Node_CreateRoom{
		RoomId:   roomId,
		State:    1,
		RoomAddr: rms.etcdAddr,
	}
	_, ok := rms.GetRoom(roomId)
	if ok {
		return
	}
	room := rooms.NewRoom(req.GetConfId(), roomId, req.GetMatchMember())
	rms.AddRoom(room)
	go room.Start(context.Background())
	return rsp, nil
}

func (rms *RoomMixServer) EnterRoom(_ctx context.Context, req *pb.Req_Node_EnterRoom) (*pb.Rsp_Node_EnterRoom, error) {
	roomId := req.GetRoomId()
	rsp := &pb.Rsp_Node_EnterRoom{
		State: 1,
	}
	room, ok := rms.GetRoom(roomId)
	if !ok {
		rsp.State = 2
		return rsp, nil
	}
	room.EnterPlayerActor(req.GetPlayerId(), req.GetActor())
	return rsp, nil
}

func (rms *RoomMixServer) QuitRoom(_ctx context.Context, req *pb.Req_Node_QuitRoom) (*pb.Rsp_Node_QuitRoom, error) {
	roomId := req.GetRoomId()
	rsp := &pb.Rsp_Node_QuitRoom{
		State: 1,
	}
	room, ok := rms.GetRoom(roomId)
	if !ok {
		rsp.State = 2
		return rsp, nil
	}
	room.RemovePlayer(req.GetPlayerId())
	return rsp, nil
}
