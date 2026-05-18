package model_game

import (
	"fmt"
	"g7/common/protos/pb"
)

func (p *Player) GetSeverAndPlayerId() string {
	return fmt.Sprintf("%d_%d", p.ServerId, p.PlayerId)
}

func (p *Player) ToBattleActor() *pb.BattleActor {
	var battleActor = &pb.BattleActor{}
	p.RunInActor(func() {
		battleActor.ActorId = p.PlayerId
		battleActor.Name = p.Nickname

	})
	return battleActor
}

func (p *Player) ToClientInfo() *pb.GamePlayerInfo {
	var gamePlayerInfo = &pb.GamePlayerInfo{}
	p.RunInActor(func() {
		gamePlayerInfo.PlayerId = p.PlayerId
		gamePlayerInfo.ServrId = p.ServerId
		gamePlayerInfo.NickName = p.Nickname
	})
	return gamePlayerInfo
}
