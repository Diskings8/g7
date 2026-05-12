package roommanager_server

import (
	"context"
	"fmt"
	"g7/common/etcd"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"sync"
	"time"
)

var GRoomManagerServer = &RoomManagerServer{}

type RoomManagerServer struct {
	pb.UnimplementedRoomManagerNodeServiceServer
	rw         sync.RWMutex
	comMonitor *etcd.ComMonitor
}

func (rms *RoomManagerServer) Init() {
	rms.comMonitor = etcd.NewComMonitor(etcd.GetRoomRpcPrefix())
	rms.comMonitor.Run()
	rms.InitMasterRom()
}

func (rms *RoomManagerServer) InitMasterRom() {
	go func() {
		time.Sleep(5 * time.Second)
		req := &pb.Req_Node_CreateRoom{
			RoomType:    1,
			ConfId:      101,
			MatchMember: nil,
			MatchId:     "master_room",
		}
		rsp, err := rms.AuthorityCreateRoom(context.Background(), req)
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		RoomKey := redisx.MakeRoomMasterKey(req.GetConfId(), req.GetRoomType())
		RoomValue := redisx.MakeRoomMasterValue(rsp.GetRoomId(), rsp.GetRoomAddr())
		err = redisx.SetStringKey(RoomKey, RoomValue, -1)
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}()
}

func (rms *RoomManagerServer) AuthorityCreateRoom(_ctx context.Context, req *pb.Req_Node_CreateRoom) (*pb.Rsp_Node_CreateRoom, error) {
	rsp := &pb.Rsp_Node_CreateRoom{
		State: 1,
	}
	addr, ok := rms.comMonitor.GetRandServerAddr()
	if !ok {
		rsp.State = 2
		logger.Log.Warn(fmt.Sprintf("RoomManagerServer.GetRandServerAddr:%+v", ok))
		return rsp, nil
	}
	cli, err := protocol.NewRoomNodeClient(addr)
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.State = 2
		return rsp, nil
	}
	rspNode, err := cli.CreateRoom(context.Background(),
		&pb.Req_Node_CreateRoom{RoomType: req.RoomType, MatchMember: req.GetMatchMember(),
			ConfId: req.ConfId, MatchId: req.GetMatchId()})
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.State = 2
		return rsp, nil
	}
	//logger.Log.Warn(fmt.Sprintf("RoomManagerServer.AuthorityCreateRoom:%+v", rspNode))
	rsp.RoomId = rspNode.RoomId
	rsp.RoomAddr = rspNode.RoomAddr
	return rsp, nil
}
