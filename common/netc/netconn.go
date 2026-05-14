package netc

import (
	"g7/common/protocol"
	"g7/common/protos/pb"
)

type NetConnInterface interface {
	Close() error
	ReadFromConn() (*protocol.Packet, error)
	WriteToConn(seq uint32, message *pb.GameMessage) error
	RemoteAddr() string
}
