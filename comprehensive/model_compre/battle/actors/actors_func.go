package actors

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle"
)

func (a *Actor) ToProto(isFull bool) *pb.ActorState {
	result := &pb.ActorState{
		ActorId: a.id,
		IsMove:  a.States.IsMoving,
	}
	result.Move = a.pos.ToProto()
	return result
}

func (a *Actor) TakeEffect(srcId int64, score float64) {
	if a.IsDead {
		return
	}
	a.Attributes.Cost(battle.AttributesHp, score)
	if a.Attributes.Hp <= 0 {
		a.IsDead = true
		a.KillerId = srcId
	}
}
