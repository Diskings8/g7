package general_system_game

import (
	"context"
	"fmt"
	"g7/common/etcd"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/game/const_game"
	"g7/game/manager_game"
	"g7/game/model_game"
)

var GMatchSystem = &matchSystem{}

type matchSystem struct {
}

func (this *matchSystem) Init() {
	manager_game.GISystemManager.Register(const_game.General_MatchSystem, GMatchSystem)
}

func (this *matchSystem) GetName() string {
	return "general_match_system"
}

func (this *matchSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {

}

func (this *matchSystem) OnEnterGame(Player *model_game.Player) {

}

func (this *matchSystem) StartMatch(Player *model_game.Player) {
	kvL, err := etcd.GetMatchServersList()
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	if len(kvL) == 0 {
		logger.Log.Error("not match server list")
		return
	}
	severAddr := kvL[0].V
	cli, err := protocol.NewMatchNodeClient(severAddr)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	rsp, err := cli.StartMatch(context.Background(), &pb.Req_Node_NewMatch{
		PlayerId: Player.PlayerId,
		ServerId: Player.ServerId,
		Score:    1000,
	})
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	fmt.Println(rsp)
	return
}
