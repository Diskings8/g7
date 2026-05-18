package mix_server

import (
	"context"
	"fmt"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/gateway/global_gateway"
	"google.golang.org/grpc"
	"log"
	"net"
)

func RunGrpcServer(ctx context.Context, lis net.Listener, etcdGrpcAddr string) {
	grpcServer := grpc.NewServer()
	pb.RegisterGatewayNodeServiceServer(grpcServer, &GatewayMixServer{etcdGrpcAddr: etcdGrpcAddr})
	go grpcServer.Serve(lis)
	go func() {
		select {
		case <-ctx.Done():
			grpcServer.Stop()
		}
	}()
	return
}

func (gms *GatewayMixServer) GetConnCount(_ctx context.Context, req *pb.Req_Node_ConnCount) (*pb.Rsp_Node_ConnCount, error) {
	return &pb.Rsp_Node_ConnCount{Count: global_gateway.GetConnCount()}, nil
}

func (gms *GatewayMixServer) ConnToRoom(_ctx context.Context, req *pb.Req_Node_MakeConnToRoom) (*pb.Rsp_Node_MakeConnToRoom, error) {
	rsp := &pb.Rsp_Node_MakeConnToRoom{State: 2}
	//logger.Log.Info(fmt.Sprintf("RoomManagerServer.ConnToRoom:%+v", req))
	sess := global_gateway.GConnSessionMap.FindSessionByPlayerId(req.PlayerId)
	if sess == nil {
		logger.Log.Info(fmt.Sprintf("RoomManagerServer.ConnToRoom session emopty"))
		return rsp, nil
	}
	stream, err := gms.connectToRoomServer(req.GetRoomAddr())
	if err != nil {
		log.Printf("连接房间服失败: %v:%+v", err, req)
		return rsp, nil
	}
	//logger.Log.Info("get  connect")
	roomId := req.GetRoomId()
	sess.SetRoomStream(stream, roomId)
	go sess.RunGoRoutineToRecvFromRoom()
	go sess.RunGoRoutineToSendToRoom()
	return &pb.Rsp_Node_MakeConnToRoom{State: 1}, nil
}
