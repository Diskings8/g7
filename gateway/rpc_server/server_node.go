package rpc_server

import (
	"context"
	"g7/common/protos/pb"
	"g7/gateway/global_gateway"
	"google.golang.org/grpc"
	"net"
)

type GatewayNodeServer struct {
	pb.UnimplementedGatewayNodeServiceServer
}

func (s *GatewayNodeServer) GetConnCount(_ctx context.Context, req *pb.Req_Node_ConnCount) (*pb.Rsp_Node_ConnCount, error) {
	return &pb.Rsp_Node_ConnCount{Count: global_gateway.GetConnCount()}, nil
}

func RunGrpcServer(lis net.Listener) {
	grpcServer := grpc.NewServer()
	pb.RegisterGatewayNodeServiceServer(grpcServer, &GatewayNodeServer{})
	grpcServer.Serve(lis)
}
