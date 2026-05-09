package interfaces

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/common_battle"
	"time"
)

type Actor interface {
	ID() int64
	Type() battle.ActorType // Player, Monster, etc.
	Update(delta time.Duration, world World)
	ToProto(bool) *pb.ActorState
	IsDirty() bool
	ClearDirty()
	Pos() common_battle.Vector3D
	// 处理输入（仅玩家用）
	AcceptInput(input actoractions.ActorAction)
}
