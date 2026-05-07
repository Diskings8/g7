package match_server

import (
	"context"
	"g7/common/protos/pb"
	"g7/comprehensive/manager_system"
)

type MatchServer struct {
	pb.UnimplementedMatchNodeServiceServer
}

func (ms *MatchServer) StartMatch(_ctx context.Context, req *pb.Req_Node_NewMatch) (*pb.Rsp_Node_NewMatch, error) {
	rsp := &pb.Rsp_Node_NewMatch{
		State: 1,
	}
	matcherInfo := manager_system.GMatchManager.NewMatcher(req.GetWaiter())
	err := manager_system.GMatchManager.StarMatch(matcherInfo)
	if err != nil {
		rsp.State = 2
		rsp.ErrReason = err.Error()
		return rsp, nil
	}
	rsp.ExpectedWaitTime = 3 * 60
	return rsp, nil
}
