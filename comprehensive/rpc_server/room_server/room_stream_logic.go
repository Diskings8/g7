package room_server

import (
	"context"
	"errors"
	"fmt"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/rooms"
	"github.com/golang/protobuf/proto"
)

func (rms *RoomMixServer) Stream(stream pb.RoomStreamService_StreamServer) (err error) {
	_, StreamCancel := context.WithCancel(stream.Context())
	var playerStream *rooms.PlayerStream
	for {
		pkt, err := stream.Recv()
		if err != nil {
			if playerStream != nil {
				playerStream.Quit()
			}
			break
		}
		if pb.MsgID(pkt.MsgId) == pb.MsgID_MSG_AUTH {
			playerStream = rms.handleAuth(pkt.GetBody(), stream, StreamCancel)
			continue
		}

		if playerStream == nil {
			logger.Log.Warn(fmt.Sprintf("%d,not have auth playerStream", pkt.MsgId))
			err = errors.New("not have auth playerStream")
			break
		}
		playerStream.Recv(pkt)
	}
	StreamCancel()

	return nil
}

func (rms *RoomMixServer) handleAuth(msgBody []byte, stream pb.RoomStreamService_StreamServer, StreamCancel func()) *rooms.PlayerStream {
	req := pb.Req_AuthClientToRoom{}
	err := proto.Unmarshal(msgBody, &req)
	if err != nil {
		return nil
	}
	room, ok := rms.roomMaps[req.GetRoomId()]
	if !ok {
		logger.Log.Warn(fmt.Sprintf("%s,not have room", req.GetRoomId()))
		return nil
	}
	room.SetPlayerStream(req.GetPlayerID(), stream)
	ps, ok := room.GetPlayerStream(req.GetPlayerID())
	return ps
}
