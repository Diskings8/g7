package rooms

import (
	"g7/common/protos/pb"
	"time"
)

type World struct {
}

func (w *World) Step(tick time.Duration) {

}

func (w *World) GetSnapshot() *pb.GameMessage {
	return &pb.GameMessage{}
}
