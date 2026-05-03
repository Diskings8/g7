package rpc_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"g7/common/logger"
	"g7/common/mqc/mq_topic"
	"g7/common/protos/pb"
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
	globalSequence uint64
}

// Stream 实现双向流方法
func (s *GameStreamServer) Stream(stream pb.GameStreamService_StreamServer) (err error) {
	//log.Println("玩家流连接建立")
	var player *model_game.Player
	_, StreamCancel := context.WithCancel(stream.Context())
	streamId := s.NewSteamId()

	// 循环接收网关转发的客户端消息
	for {
		pkt, err := stream.Recv()
		if err != nil {
			//logger.Log.Warn(fmt.Sprintf("Recv ： %s", err))
			break
		}
		if pb.MsgID(pkt.MsgId) == pb.MsgID_MSG_AUTH {
			player = s.handleAuth(pkt.GetBody(), stream, streamId, StreamCancel)
			continue
		}

		if player == nil {
			logger.Log.Warn(fmt.Sprintf("%d,not have auth player", pkt.MsgId))
			err = errors.New("not have auth player")
			break
		}
		if player.StreamId != streamId {
			// 不是同一个流了，需要断开
			logger.Log.Warn(fmt.Sprintf("%d,another conn login", player.PlayerId))
			break
		}
		//logger.Log.Info(fmt.Sprintf("%s", pb.MsgID(pkt.MsgId)))
		// 这里写你的游戏逻辑：根据 msg_id 处理 body
		player.RunInActor(func() {
			if !s.isAllow(player) {
				logger.Log.Info(fmt.Sprintf("%d not allow", player.PlayerId))
				return
			}
			// 更新心跳
			player.LastHearBeatTime = time.Now()
			// 处理逻辑
			rsp := s.handleGameMessageLogic(pb.MsgID(pkt.MsgId), pkt.Body, player)
			if rsp != nil {
				player.SendMessage(pb.MsgID(pkt.GetMsgId()), rsp)
			}
			//处理mq日志
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

func (s *GameStreamServer) handleAuth(data []byte, stream pb.GameStreamService_StreamServer, streamId uint64, cancelFunc func()) (player *model_game.Player) {
	req := pb.Req_AuthClientToGame{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil
	}

	// 重连
	if val := global_game.GPlayerMaps.GetPlayer(req.GetPlayerID()); val != nil {
		player = val
		player.StreamId = streamId

		player.StreamConn = stream
		player.StreamCancelFunc = cancelFunc
		player.IsOnline = true
		player.OfflineAt = time.Time{}
		//logger.Log.Info(fmt.Sprintf("玩家 %d 重连成功", player.PlayerId))
		return
	}
	// 新上线 从redis获取缓存加载
	playerId := req.PlayerID
	serverId := req.ServerID
	playerDao, err := global_game.GPlayerCache.GetPlayerCache(serverId, playerId)
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
	player.StreamId = streamId
	manager_game.NewPlayerBase(player, stream, cancelFunc)

	manager_game.GISystemManager.LoadData(playerDao, player)
	global_game.GPlayerMaps.SetPlayer(player.PlayerId, player)
	global_game.GPlayerMaps.RegisterRedisLoginKey(player)
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

func (s *GameStreamServer) NewSteamId() uint64 {
	return atomic.AddUint64(&s.globalSequence, 1)
}
