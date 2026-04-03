package handle_stream

import (
	"g7/common/protos/pb"
	"g7/game/model_game"
)

func HandleLogic(MsgId pb.MsgID, data []byte, player *model_game.Player) (rsp any) {
	switch MsgId {
	case pb.MsgID_MSG_ENTER_GAME:
		rsp = handle_MSG_ENTER_GAME(data, player)

	}

	return
}

func handle_MSG_ENTER_GAME(data []byte, player *model_game.Player) (rsp any) {
	rsp = &pb.Rsp_LoginGame{Result: true}
	return
}
