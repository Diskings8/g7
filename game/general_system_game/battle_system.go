package general_system_game

import (
	"context"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/game/const_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"github.com/golang/protobuf/proto"
)

var GBattleSystem = &battleSystem{}

type battleSystem struct {
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_BattleSystem, GBattleSystem)
}

func (this *battleSystem) Init() {
}

func (this *battleSystem) GetName() string {
	return "general_bag_system"
}

func (this *battleSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {
}

func (this *battleSystem) OnEnterGame(Player *model_game.Player) {}

func (this *battleSystem) ReqToEnterScene(reqData []byte, Player *model_game.Player) (rsp *pb.Rsp_EnterRoom) {
	req := &pb.Req_EnterRoom{}
	_ = proto.Unmarshal(reqData, req)
	rsp = &pb.Rsp_EnterRoom{Result: true}

	nodeReq := pb.Req_Node_EnterRoom{
		PlayerId: Player.PlayerId,
		RoomId:   Player.RoomId,
		Actor:    Player.ToBattleActor(),
	}
	nodeAddr := Player.RoomAddr
	cli, err := protocol.NewRoomNodeClient(nodeAddr)
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.Result = false
		return
	}
	_, err = cli.EnterRoom(context.Background(), &nodeReq)
	if err != nil {
		logger.Log.Error(err.Error())
		rsp.Result = false
		return
	}
	return
}
