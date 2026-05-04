package general_system_game

import (
	"context"
	"fmt"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/snowflakes"
	"g7/game/const_game"
	"g7/game/global_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"github.com/golang/protobuf/proto"
	"time"
)

var GOrderSystem = &orderService{}

type orderService struct {
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_OrderSystem, GOrderSystem)
}

func (this *orderService) Init() {

}

func (this *orderService) LoadData(PlayerDao *model_game.PlayerDao, Player *model_game.Player) {

}
func (this *orderService) OnEnterGame(Player *model_game.Player) {

}
func (this *orderService) GetName() string {
	return "General_OrderSystem"
}

func (this *orderService) CreateOrder(reqData []byte, Player *model_game.Player) any {
	req := &pb.Req_CreateOrder{}
	_ = proto.Unmarshal(reqData, req)

	playerId := Player.PlayerId
	severId := Player.ServerId

	go func() {
		defer func() {
			if cover := recover(); cover != nil {
				logger.Log.Error(cover.(string))
			}
		}()
		rsp := &pb.Rsp_CreateOrder{}
		ctx := context.Background()
		lockKey := fmt.Sprintf(redisx.ShopLockKeyPrefix, severId, playerId)
		locked, err := redisx.GetClient().SetNX(ctx, lockKey, "1", 5*time.Second).Result()
		if err != nil || !locked {
			rsp.ErrReason = "操作太频繁，请稍后再试"
			return
		}
		defer redisx.GetClient().Del(ctx, lockKey)
		orderId := this.generateOrderNo()
		order := &model_common.GameOrder{
			OrderNo:     orderId,
			PlayerID:    playerId,
			ServerID:    severId,
			ProductID:   req.GetProductId(),
			ProductName: "product.Name",
			ProductType: 1,
			Price:       12,
			Currency:    "CNY",
			PayType:     0,
			PayAmount:   0,
			Status:      globals.OrderStatusPending,
			CreateTime:  time.Now().Unix(),
		}

		err = global_game.GGlobalDB.Insert(order)
		if err != nil {
			logger.Log.Error(err.Error())
			rsp.ErrReason = "订单生成失败"
			return
		}
		rsp.OrderId = orderId
		Player.RunInActor(func() {
			Player.SendMessage(pb.MsgID_MSG_Rsp_CreateOrder, rsp)
			this.makeLoginLog(Player, orderId)
		})
	}()
	return nil
}

func (this *orderService) generateOrderNo() string {
	return snowflakes.GenStringID()
}

func (this *orderService) makeLoginLog(player *model_game.Player, orderId string) {
	ld := model_common.ActionLog{
		BaseLog:      model_common.BaseLog{ServerId: player.ServerId, EventType: globals.ActionEventLogin, CreateTime: time.Now().Unix()},
		PlayerID:     player.PlayerId,
		Action:       "CreateOrder",
		Reason:       "",
		CostItem:     nil,
		CostCurrency: nil,
		GainItem:     nil,
		GainCurrency: nil,
		Ext:          orderId,
	}
	player.ActionLogs = append(player.ActionLogs, &ld)
}

func (this *orderService) GrantRewards(rewards map[int32]int64, player *model_game.Player) {

}
