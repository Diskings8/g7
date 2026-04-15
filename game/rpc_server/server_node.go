package rpc_server

import (
	"context"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/protos/pb"
	"g7/common/snowflakes"
	"g7/game/general_system_game"
	"g7/game/global_game"
	"g7/game/manager_game"
	"g7/game/model_game"
)

type GameNodeServer struct {
	pb.UnimplementedGameNodeServiceServer
}

func (s *GameNodeServer) LoginNodeCreatePlayer(_ctx context.Context, req *pb.Req_Node_CreatePlayer) (*pb.Rsp_Node_CreatePlayer, error) {
	//logger.Log.Info("pb.Req_Node_CreatePlayer")
	player := &model_game.Player{
		UserId:   req.GetUserID(),
		PlayerId: snowflakes.GenUID(),
		ServerId: req.GetServerID(),
		Nickname: req.GetNickname(),
	}
	daoD := player.ToDao(globals.SaveDataKindCornDb)
	// 初始化各个系统的数据
	manager_game.GISystemManager.LoadData(daoD.SaveData, player)

	rsp := &pb.Rsp_Node_CreatePlayer{
		PlayerID: player.PlayerId,
		ServerID: player.ServerId,
		Nickname: player.Nickname,
		UserID:   player.UserId,
	}
	err := global_game.GGameDB.Insert(daoD.SaveData)
	if err != nil {
		logger.Log.Error(err.Error())
		rsp = &pb.Rsp_Node_CreatePlayer{}
		rsp.State = 500
	} else {
		rsp.State = 200
		indexPlayer := model_common.GlobalPlayerIndex{
			PlayerId: player.PlayerId,
			UserID:   player.UserId,
			ServerID: player.ServerId,
			Nickname: player.Nickname,
		}
		_ = global_game.GGlobalDB.Insert(&indexPlayer)
	}
	return rsp, nil
}

func (s *GameNodeServer) LoginNodeOrderPaid(_ctx context.Context, req *pb.Req_Node_OrderPaid) (*pb.Rsp_Node_OrderPaid, error) {
	req.GetOrderId()
	order := &model_common.GameOrder{}
	_ = global_game.GGlobalDB.FindOne(order, map[string]interface{}{"order_no": req.OrderId})
	if order.Status != globals.OrderStatusPaid {
		return &pb.Rsp_Node_OrderPaid{
			State: 0,
		}, nil
	}
	order.Status = globals.OrderStatusProcessing
	_ = global_game.GGlobalDB.Insert(order)

	reward := s.GenOrderItems()
	player := global_game.GPlayerMaps.GetPlayer(req.GetPlayerID())
	if player == nil {
		return &pb.Rsp_Node_OrderPaid{
			State: 0,
		}, nil
	}

	player.RunInActor(func() {
		general_system_game.GOrderSystem.GrantRewards(reward, player)
	})
	order.Status = globals.OrderStatusCompleted
	_ = global_game.GGlobalDB.Insert(order)
	return &pb.Rsp_Node_OrderPaid{State: 1}, nil
}

func (s *GameNodeServer) GenOrderItems() map[int32]int64 {
	return make(map[int32]int64)
}
