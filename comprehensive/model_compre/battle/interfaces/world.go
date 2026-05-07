package interfaces

import "time"

type World interface {
	Start()                    // 启动世界循环
	Stop()                     // 停止世界
	AddActor(actor Actor)      // 添加单位
	RemoveActor(actorID int64) // 移除单位
	Tick(delta time.Duration)  // 单帧逻辑
}
