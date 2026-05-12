package battle

import "time"

const (
	FrameRate = 30 // 30帧每秒
	FrameTime = 1000 / FrameRate * time.Millisecond
)

type ActorType int32

const (
	ActorTypePlayer ActorType = iota + 10
)
const (
	ActorTypeMonster ActorType = iota + 100
)

const (
	AttributesHp = iota + 1
	AttributesMp
)
