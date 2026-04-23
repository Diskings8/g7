package handle_stream

import (
	"encoding/json"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/protos/pb"
	"g7/common/structs"
	"g7/common/utils"
	"g7/game/general_system_game"
	"g7/game/global_game"
	"g7/game/model_game"
	"github.com/golang/protobuf/proto"
	"strings"
	"time"
)

func handleGmCmd(reqD []byte, player *model_game.Player) any {
	if globals.IsProd() {
		return nil
	}
	rsp := &pb.Rsp_RunGm{State: 1}
	req := &pb.Req_RunGm{}
	err := proto.Unmarshal(reqD, req)
	if err != nil {
		logger.Log.Info(err.Error())
	}
	cmds := strings.Split(req.GetCmd(), " ")
	if len(cmds) <= 0 {
		return rsp
	}
	switch cmds[0] {
	case "add":
		switch cmds[1] {
		case "item":
			k := utils.StringToInit32(cmds[2])
			v := utils.StringToInit64(cmds[3])
			general_system_game.GBagSystem.GainAndConsumption([]structs.KInt32VInt64Bind{{k, v, 1}}, nil, "gm add", player)
		}
	case "kick":
		player.Close()
		global_game.GPlayerMaps.DelOnePlayerById(player.PlayerId)

	case "del":
		switch cmds[1] {
		case "item":
			k := utils.StringToInit32(cmds[2])
			v := utils.StringToInit64(cmds[3])
			general_system_game.GBagSystem.GainAndConsumption(nil, []structs.KInt32VInt64Bind{{k, v, 0}}, "gm del", player)
		}
	case "pay":
		k := utils.StringToInit32(cmds[1])
		d, _ := json.Marshal(&pb.Req_CreateOrder{ProductId: k})
		r := general_system_game.GOrderSystem.CreateOrder(d, player)
		if r != nil {
			rsp.Ext = r.(*pb.Rsp_CreateOrder).OrderId
		}
	case "mailAll":
		general_system_game.GMailSystem.SendDefaultSystemTypeMail("test", "test content", []model_common.Attachment{{ItemID: 1032, Count: 1, Bind: 0}}, time.Now().Unix(), 3, "tester")
	case "match":
		general_system_game.GMatchSystem.StartMatch(player)
	}
	player.RedisReWrite(globals.SaveDataKindCornCache)
	return rsp
}
