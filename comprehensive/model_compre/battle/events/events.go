package events

type Event struct {
	EventType int32
	CasterId  int64
	TargetId  int64
}
