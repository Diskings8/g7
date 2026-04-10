package rpc_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"g7/common/logger"
	"g7/common/mqc/mq_topic"
	"g7/common/protos/pb"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/db_game"
	"g7/game/global_game"
	"g7/game/handle_stream"
	"g7/game/manager_game"
	"g7/game/model_game"
	"sync/atomic"
	"time"
)

// GameStreamServer 实现接口
type GameStreamServer struct {
	pb.UnimplementedGameStreamServiceServer
}

// Stream 实现双向流方法
func (s *GameStreamServer) Stream(stream pb.GameStreamService_StreamServer) (err error) {
	//log.Println("玩家流连接建立")
	var player *model_game.Player
	_, StreamCancel := context.WithCancel(stream.Context())

	// 申请流id
	streamId := global_game.NewStreamID()

	// 循环接收网关转发的客户端消息
	for {
		pkt, err := stream.Recv()
		if err != nil {
			//log.Printf("流断开: %v", err)
			if player != nil {
				s.handleStreamDisconnect(player, streamId)
			}
			break
		}
		if pb.MsgID(pkt.MsgId) == pb.MsgID_MSG_AUTH {
			player = s.handleAuth(pkt.GetBody(), stream, StreamCancel)
			player.StreamID = streamId
			continue
		}

		if player == nil {
			//logger.Log.Warn(fmt.Sprintf("%s,not have auth player", pkt.MsgId))
			err = errors.New("not have auth player")
			break
		}
		//log.Printf("收到消息: msg_id=%d, body_len=%d", pkt.MsgId, len(pkt.Body))

		// 这里写你的游戏逻辑：根据 msg_id 处理 body
		player.RunInActor(func() {
			if s.isAllow(player) {
				return
			}
			rsp := s.handleGameMessageLogic(pb.MsgID(pkt.MsgId), pkt.Body, player)
			if rsp != nil {
				player.SendMessage(pb.MsgID(pkt.GetMsgId()), rsp)
			}
			s.handleGameMQCreate(player)
		})
	}
	StreamCancel()

	return nil
}

func (s *GameStreamServer) handleGameMessageLogic(msgID pb.MsgID, data []byte, player *model_game.Player) (rsp any) {
	return handle_stream.HandleLogic(msgID, data, player)
}

func (s *GameStreamServer) handleGameMQCreate(player *model_game.Player) {
	valL := player.GetAllActionLogs()
	for _, v := range valL {
		global_game.GGlobalMQ.ProduceMessage(mq_topic.MakeGameActionTopicKey(), v)
	}
}

func (s *GameStreamServer) handleStreamDisconnect(p *model_game.Player, streamId uint64) {
	if streamId < p.StreamID {
		return // 直接忽略，不做任何操作
	}

	if !p.IsDisconnecting.CompareAndSwap(false, true) {
		// 已有断线流程执行中 → 本次直接丢弃，不处理
		return
	}

	// 函数退出时，释放状态
	defer p.IsDisconnecting.Store(false)
	// 标记离线
	p.MarkOffLine()
	// 2. 如果已有旧定时器，先取消
	if p.DisconnectCancelTimer != nil {
		p.DisconnectCancelTimer()
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(const_game.TcpCloseWaitTime))
	p.DisconnectCancelTimer = cancel

	go func() {
		// 等待 离线上限
		select {
		case <-time.After(utils.DisConnectMaxTimeLimit):
			// 超时 → 真正下线！
			p.QuitChan <- struct{}{} // 确定下线
			global_game.GPlayerMaps.DelPlayer(p.PlayerId)
			//log.Printf("玩家 %d 超时未重连 → 正式下线", p.PlayerId)
		case <-ctx.Done():
			// 玩家重连了 → 取消下线
			//log.Printf("玩家 %d 重连成功 → 取消离线", p.PlayerId)
		}
		//logger.Log.Info(fmt.Sprintf("handleStreamDisconnect routine end: %d", p.PlayerId))
	}()

}

func (s *GameStreamServer) handleAuth(data []byte, stream pb.GameStreamService_StreamServer, cancelFunc func()) (player *model_game.Player) {
	req := pb.Req_AuthClientToGame{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil
	}

	// 重连
	if val := global_game.GPlayerMaps.GetPlayer(req.GetPlayerID()); val != nil {
		if val.DisconnectCancelTimer != nil {
			val.DisconnectCancelTimer()
		}
		player = val

		player.StreamConn = stream
		player.StreamCancelFunc = cancelFunc
		player.IsOnline = true
		player.OfflineAt = time.Time{}
		//logger.Log.Info(fmt.Sprintf("玩家 %d 重连成功", player.PlayerId))
		return
	}
	// 新上线 从redis获取缓存加载
	playerId := req.PlayerID
	playerDao, err := global_game.GPlayerCache.GetPlayerCache(playerId)
	if err != nil {
		//logger.Log.Info(fmt.Sprintf("玩家 %d Redis缓存加载失败，降级到MySQL: %v", playerId, err))
		// 2.2 Redis加载失败，兜底从MySQL加载冷数据
		playerDao, err = db_game.GetPlayerByID(playerId)
		if err != nil {
			logger.Log.Info(fmt.Sprintf("玩家 %d MySQL加载失败: %v", playerId, err))
			return nil
		}
		// 2.3 加载成功后，回填Redis（下次登录走缓存）
		if err := global_game.GPlayerCache.SetPlayerCache(playerDao); err != nil {
			logger.Log.Info(fmt.Sprintf("玩家 %d 回填Redis失败: %v", playerId, err))
			// 这里不阻塞，不影响玩家登录，只打日志告警
		}
	}
	player = playerDao.TomSimplePlayer()
	manager_game.NewPlayerBase(player, stream, cancelFunc)

	manager_game.GISystemManager.LoadData(playerDao, player)
	global_game.GPlayerMaps.SetPlayer(player.PlayerId, player)
	//
	manager_game.OnLineRunning(player)

	return
}

func (s *GameStreamServer) isAllow(p *model_game.Player) bool {
	now := time.Now().Unix()

	// 如果新的一秒，重置
	if atomic.LoadInt64(&p.LimitLastReqTime) != now {
		atomic.StoreInt32(&p.LimitReqCount, 1)
		atomic.StoreInt64(&p.LimitLastReqTime, now)
		return true
	}

	// 计数+1，判断是否超限
	return atomic.AddInt32(&p.LimitReqCount, 1) <= 30
}
