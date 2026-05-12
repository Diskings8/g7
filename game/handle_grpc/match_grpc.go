package handle_grpc

import (
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/utils"
	"g7/game/general_system_game"
	"g7/game/global_game"
)

func HandleMatchSuccess(req *pb.Req_Node_MatchSuccess) {
	for _, v := range req.GetMatchMember() {
		playerIdStr := redisx.GetPlayerLoginValueIndex(v, redisx.RedisPlayerLoginValIndexPlayerId)
		playerId := utils.StringToInt64(playerIdStr)
		player := global_game.GPlayerMaps.GetPlayer(playerId)
		//logger.Log.Info(fmt.Sprintf("HandleMatchSuccess:%+v", req))
		// todo
		general_system_game.GBattleSystem.RoomDataChange(req.GetRoomId(), req.GetRoomAddr(), player)
		player.SendMessage(pb.MsgID_MSG_Notify_MatchSuccess, &pb.Notify_MatchSuccess{Result: true, Reason: "匹配成功"})
	}

}
