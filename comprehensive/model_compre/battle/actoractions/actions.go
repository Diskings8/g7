package actoractions

import "g7/common/protos/pb"

type ActorAction struct {
	ActorId int64
	Action  *pb.GameMessage
}
