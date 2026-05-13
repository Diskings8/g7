package handle_stream

import (
	"g7/common/protos/pb"
	"g7/game/general_system_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"github.com/golang/protobuf/proto"
)

func HandleLogic(MsgId pb.MsgID, data []byte, player *model_game.Player) (rsp proto.Message) {
	switch MsgId {
	case pb.MsgID_MSG_Req_EnterGame:
		rsp = handleMsgEnterGame(data, player)
	case pb.MsgID_MSG_Req_CreateOrder:
		rsp = handleMsgCreateOrder(data, player)
	case pb.MsgID_MSG_GM_Cmd:
		rsp = handleGmCmd(data, player)
	case pb.MsgID_MSG_Req_EnterScene:
		rsp = handleMsgEnterScene(data, player)
	case pb.MsgID_MSG_Req_StartMatch:
		rsp = handleMsgStartMatch(data, player)
	}
	return
}

func handleMsgEnterGame(req []byte, player *model_game.Player) (rsp proto.Message) {

	player.RunInActor(func() {
		manager_game.GResetSystemManager.AllReset(player)
		manager_game.GISystemManager.OnEnterGame(player)
	})

	rsp = &pb.Rsp_LoginGame{Result: true}
	return
}

func handleMsgCreateOrder(req []byte, player *model_game.Player) (rsp proto.Message) {
	return general_system_game.GOrderSystem.CreateOrder(req, player)
}

func handleMsgEnterScene(req []byte, player *model_game.Player) (rsp proto.Message) {
	return general_system_game.GBattleSystem.ReqToEnterScene(req, player)
}

func handleMsgStartMatch(req []byte, player *model_game.Player) (rsp proto.Message) {
	reqD := &pb.Req_StartMatch{}
	proto.Unmarshal(req, reqD)
	return general_system_game.GMatchSystem.StartMatch(reqD, player)
}
