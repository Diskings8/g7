package rpc_server

import (
	"context"
	"g7/common/protos/pb"
	"g7/comprehensive/manager_system"
)

type MatchServer struct {
	pb.UnimplementedMatchNodeServiceServer
}

func (ms *MatchServer) StartMatch(_ctx context.Context, req *pb.Req_Node_NewMatch) (*pb.Rsp_Node_NewMatch, error) {
	rsp := &pb.Rsp_Node_NewMatch{}
	err := manager_system.GMatchManager.NewMatcher(req.GetPlayerId(), req.GetServerId())
	if err != nil {
		rsp.State = 0
	}
	return rsp, nil
}
