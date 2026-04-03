package rpc_server

import (
	"context"
	"encoding/json"
	"errors"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/game/global_game"
	"g7/game/handle_stream"
	"g7/game/model_game"
	"log"
	"time"
)

// GameStreamServer 实现接口
type GameStreamServer struct {
	pb.UnimplementedGameStreamServiceServer
}

// Stream 实现双向流方法
func (s *GameStreamServer) Stream(stream pb.GameStreamService_StreamServer) error {
	//log.Println("玩家流连接建立")
	var player *model_game.Player
	// 循环接收网关转发的客户端消息
	for {
		pkt, err := stream.Recv()
		if err != nil {
			//log.Printf("流断开: %v", err)
			if player != nil {
				s.handleStreamDisconnect(player)
			}
			return err
		}
		if pb.MsgID(pkt.MsgId) == pb.MsgID_MSG_AUTH {
			player = s.handleAuth(pkt.GetBody(), stream)
		}

		if player != nil {
			logger.Log.Warn("not have auth player")
			return errors.New("not have auth player")
		}
		//log.Printf("收到消息: msg_id=%d, body_len=%d", pkt.MsgId, len(pkt.Body))

		// 这里写你的游戏逻辑：根据 msg_id 处理 body
		rsp := s.handleGameMessageLogic(pb.MsgID(pkt.MsgId), pkt.Body, player)
		if rsp != nil {
			player.SendMessage(pb.MsgID(pkt.GetMsgId()), rsp)
		}

	}
}

func (s *GameStreamServer) handleGameMessageLogic(msgID pb.MsgID, data []byte, player *model_game.Player) (rsp any) {
	return handle_stream.HandleLogic(msgID, data, player)
}

func (s *GameStreamServer) handleStreamDisconnect(p *model_game.Player) {
	p.Online = false
	p.StreamConn = nil
	p.OfflineAt = time.Now()

	// 2. 如果已有旧定时器，先取消
	if p.CancelTimer != nil {
		p.CancelTimer()
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*3))
	p.CancelTimer = cancel

	go func() {
		// 等待 8 秒
		select {
		case <-time.After(8 * time.Second):
			// 超时 → 真正下线！
			p.Save() // 存库
			global_game.GPlayerMaps.DelPlayer(p.PlayerId)
			log.Printf("玩家 %d 超时未重连 → 正式下线", p.PlayerId)
		case <-ctx.Done():
			// 玩家重连了 → 取消下线
			log.Printf("玩家 %d 重连成功 → 取消离线", p.PlayerId)
		}
	}()
}

func (s *GameStreamServer) handleAuth(data []byte, stream pb.GameStreamService_StreamServer) (player *model_game.Player) {
	req := pb.Req_AuthClientToGame{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil
	}

	// 重连
	if val := global_game.GPlayerMaps.GetPlayer(req.GetPlayerID()); val != nil {
		if val.CancelTimer != nil {
			val.CancelTimer()
		}
		player = val

		player.StreamConn = stream
		player.Online = true
		player.OfflineAt = time.Time{}
		return
	}
	// 新上线

	// 缓存加载和数据库加载
	player = &model_game.Player{}
	player.StreamConn = stream
	player.Online = true
	player.OfflineAt = time.Time{}

	global_game.GPlayerMaps.SetPlayer(player.PlayerId, player)
	return
}
