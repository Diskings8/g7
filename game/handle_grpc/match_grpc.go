package handle_grpc

import (
	"context"
	"fmt"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/utils"
	"g7/game/global_game"
)

func HandleMatchSuccess(req *pb.Req_Node_MatchSuccess) {
	for _, v := range req.GetMatchMember() {
		playerIdStr := redisx.GetPlayerLoginValueIndex(v, redisx.RedisPlayerLoginValIndexPlayerId)
		playerId := utils.StringToInt64(playerIdStr)
		player := global_game.GPlayerMaps.GetPlayer(playerId)
		logger.Log.Info(fmt.Sprintf("HandleMatchSuccess:%+v", req))
		// todo
		player.RunInActor(func() {
			player.RoomData.RoomId = req.RoomId
			player.RoomData.RoomAddr = req.RoomAddr
		})
		gateWayAddr := player.GateWayAddr
		cli, err := protocol.NewGatewayNodeClient(gateWayAddr)
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		_, err = cli.ConnToRoom(context.Background(), &pb.Req_Node_MakeConnToRoom{PlayerId: playerId, RoomId: req.GetRoomId(), RoomAddr: req.GetRoomAddr()})
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		player.SendMessage(pb.MsgID_MSG_Notify_MatchSuccess, pb.Notify_MatchSuccess{Result: true, Reason: "匹配成功"})
	}

}
