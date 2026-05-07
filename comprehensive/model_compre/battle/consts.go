package battle

import "time"

const (
	FrameRate = 30 // 30帧每秒
	FrameTime = 1000 / FrameRate * time.Millisecond
)

type ActorType int32
