package interfaces

import "time"

type World interface {
	AddActor(actor Actor)      // 添加单位
	RemoveActor(actorID int64) // 移除单位
	Tick(delta time.Duration)  // 单帧逻辑
}
