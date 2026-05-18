package actors

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle"
	"g7/comprehensive/model_compre/battle/common_battle"
	"g7/comprehensive/model_compre/battle/interfaces"
)

type PlayerActor struct {
	Actor
	BattleInfo *pb.BattleActor
}

func NewPlayerActor(actorId int64, battleInfo *pb.BattleActor, world interfaces.World) PlayerActor {
	pa := PlayerActor{
		BattleInfo: battleInfo,
	}
	pa.Actor = NewActor(actorId, battle.ActorTypePlayer, common_battle.Vector3D{X: 280, Y: 140}, world)
	return pa
}
