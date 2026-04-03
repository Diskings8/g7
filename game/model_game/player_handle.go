package model_game

import "g7/common/protos/pb"

type PlayerHandle struct {
	StreamConn *pb.GameService_StreamServer

	PlayerData *Player
}
