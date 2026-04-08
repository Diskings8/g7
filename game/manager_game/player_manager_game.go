package manager_game

import (
	"g7/common/cronx"
	"g7/common/protos/pb"
	"g7/game/global_game"
	"g7/game/model_game"
)

var GPlayerManager = &playerManager{}

func NewPlayerBase(p *model_game.Player, StreamConn pb.GameStreamService_StreamServer, cancelFunc func()) {
	onlineData := model_game.OnlineData{}
	onlineData.Init(StreamConn, GSaveSystemManager.AsyncSaveQueue)
	p.OnlineData = onlineData
	p.StreamCancelFunc = cancelFunc
	return
}

func OnLineRunning(p *model_game.Player) {
	//关键：启动玩家专属协程
	go p.RunMainRoutine()
	go p.RunSubRoutine()
}

type playerManager struct {
}

func (p *playerManager) Init() {
	cronx.AddPer30SecondTask(func() {
		global_game.GPlayerMaps.HeartBeatCheck()
	})
}
