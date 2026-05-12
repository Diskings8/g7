package interfaces

import (
	"g7/comprehensive/model_compre/battle/common_battle"
	"g7/comprehensive/model_compre/battle/events"
	"time"
)

type World interface {
	AddActor(actor Actor)      // 添加单位
	RemoveActor(actorID int64) // 移除单位
	Tick(delta time.Duration)  // 单帧逻辑
	FindActors(src Actor, actorIds []int64, params ...any) []Actor
	AddEvent(common_battle.Vector3D, events.Event)
}
