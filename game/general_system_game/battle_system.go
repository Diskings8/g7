package general_system_game

import (
	"context"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/game/const_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"github.com/golang/protobuf/proto"
	"strings"
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
	if len(dao.GeneralD.RoomData.RoomId) != 0 {
		Player.RoomData = dao.GeneralD.RoomData
	}
}

func (this *battleSystem) OnEnterGame(Player *model_game.Player) {
	val, _ := redisx.GetKey(redisx.MakeRoomMasterKey(101, 1))
	vals := strings.Split(val, "#")
	this.RoomDataChange(vals[0], vals[1], Player)
}

func (this *battleSystem) RoomDataChange(roomId, roomAddr string, Player *model_game.Player) {
	Player.RunInActor(func() {
		Player.RoomData.RoomId = roomId
		Player.RoomData.RoomAddr = roomAddr
	})
	gateWayAddr := Player.GateWayAddr
	cli, err := protocol.NewGatewayNodeClient(gateWayAddr)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	_, err = cli.ConnToRoom(context.Background(), &pb.Req_Node_MakeConnToRoom{PlayerId: Player.PlayerId, RoomId: roomId, RoomAddr: roomAddr})
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

}

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
