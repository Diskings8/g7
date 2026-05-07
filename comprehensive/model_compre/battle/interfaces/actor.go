package interfaces

import (
	"g7/comprehensive/model_compre/battle"
	"time"
)

type Actor interface {
	ID() int64
	Type() battle.ActorType // Player, Monster, etc.
	Update(delta time.Duration, world World)
	ToProto() any
	// 处理输入（仅玩家用）
	AcceptInput(input any)
}
