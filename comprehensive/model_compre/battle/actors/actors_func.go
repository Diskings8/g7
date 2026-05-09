package actors

import "g7/common/protos/pb"

func (a *Actor) ToProto(isFull bool) *pb.ActorState {
	result := &pb.ActorState{
		ActorId: a.id,
		IsMove:  a.isMoving,
	}
	result.Move = a.pos.ToProto()
	return result
}
