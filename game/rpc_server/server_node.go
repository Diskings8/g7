package rpc_server

import (
	"context"
	"g7/common/logger"
	"g7/common/protos/pb"
)

type GameNodeServer struct {
	pb.UnimplementedGameNodeServiceServer
}

func (s *GameNodeServer) LoginNodeCreatePlayer(context.Context, *pb.Req_Node_CreatePlayer) (*pb.Rsp_Node_CreatePlayer, error) {
	logger.Log.Info("pb.Req_Node_CreatePlayer")
	rsp := &pb.Rsp_Node_CreatePlayer{
		PlayerID: 910001,
		ServerID: 91001,
		ID:       100001,
		Nickname: "Rav3n96",
		UserID:   10241,
	}
	return rsp, nil
}
