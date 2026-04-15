package handle_stream

import (
	"g7/common/protos/pb"
	"g7/game/general_system_game"
	"g7/game/manager_game"
	"g7/game/model_game"
)

func HandleLogic(MsgId pb.MsgID, data []byte, player *model_game.Player) (rsp any) {
	switch MsgId {
	case pb.MsgID_MSG_Req_EnterGame:
		rsp = handleMsgEnterGame(data, player)
	case pb.MsgID_MSG_Req_CreateOrder:
		rsp = handleMsgCreateOrder(data, player)
	}
	return
}

func handleMsgEnterGame(req []byte, player *model_game.Player) (rsp any) {

	manager_game.GResetSystemManager.AllReset(player)
	manager_game.GISystemManager.OnEnterGame(player)

	rsp = &pb.Rsp_LoginGame{Result: true}
	return
}

func handleMsgCreateOrder(req []byte, player *model_game.Player) (rsp any) {
	return general_system_game.GOrderSystem.CreateOrder(req, player)
}
