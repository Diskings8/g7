package manager_game

import (
	"g7/common/protos/pb"
	"g7/game/model_game"
)

func NewPlayerBase(p *model_game.Player, StreamConn pb.GameStreamService_StreamServer) {
	onlineData := model_game.OnlineData{}
	onlineData.Init(StreamConn, GSaveSystemManager.AsyncSaveQueue)
	p.OnlineData = onlineData
	return
}

func OnLineRunning(p *model_game.Player) {
	//关键：启动玩家专属协程
	go p.RunMainRoutine()
	go p.RunSubRoutine()
}
