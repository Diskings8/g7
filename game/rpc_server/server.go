package rpc_server

import (
	"g7/common/protos/pb"
	"log"
)

// 实现 GameServiceServer 接口
type gameServer struct {
	pb.UnimplementedGameServiceServer
}

// 实现双向流方法：StreamConnect
func (s *gameServer) StreamConnect(stream pb.GameService_StreamServer) error {
	log.Println("✅ 玩家流连接建立")

	// 循环接收网关转发的客户端消息
	for {
		pkt, err := stream.Recv()
		if err != nil {
			log.Printf("流断开: %v", err)
			return err
		}

		log.Printf("收到消息: msg_id=%d, body_len=%d", pkt.MsgId, len(pkt.Body))

		// 这里写你的游戏逻辑：根据 msg_id 处理 body
		// handleGameLogic(pkt.MsgId, pkt.Body)

		// 处理完后，把结果推回给客户端（通过同一个流）
		// stream.Send(&pb.ServerPacket{
		// 	MsgId: pkt.MsgId,
		// 	Body:  []byte("处理完成"),
		// })
	}
}
