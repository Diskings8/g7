package interfaces

import "g7/common/protos/pb"

type Room interface {
	Broadcast(*pb.GameMessage)
}
