package general_system_game

import (
	"context"
	"g7/common/etcd"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/utils"
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

func (this *matchSystem) StartMatch(req *pb.Req_StartMatch, Player *model_game.Player) (rsp *pb.Rsp_StartMatch) {
	rsp = &pb.Rsp_StartMatch{Result: true}
	kvL, err := etcd.GetMatchServersList()
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.Result = false
		rsp.Reason = err.Error()
		return
	}
	if len(kvL) == 0 {
		logger.Log.Error("not match server list")
		rsp.Result = false
		rsp.Reason = "not match server list"
		return
	}
	severAddr := kvL[0].V
	cli, err := protocol.NewMatchNodeClient(severAddr)
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.Result = false
		rsp.Reason = err.Error()
		return
	}
	waiter := pb.MatchWaiter{
		TeamId:     utils.Int64ToString(Player.PlayerId),
		Score:      1000,
		TeamLeader: Player.GetSeverAndPlayerId(),
		TeamMember: []string{Player.GetSeverAndPlayerId()},
		ConfId:     req.GetConfId(),
	}
	_, err = cli.StartMatch(context.Background(), &pb.Req_Node_NewMatch{Waiter: &waiter})
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.Result = false
		rsp.Reason = err.Error()
		return
	}
	//fmt.Println(rsp)
	return
}
