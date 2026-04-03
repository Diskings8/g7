package model_game

import (
	"encoding/json"
	"g7/common/protos/pb"
	"time"
)

type OnlineData struct {
	StreamConn  pb.GameStreamService_StreamServer
	Online      bool      // 是否在线
	OfflineAt   time.Time // 离线时间
	CancelTimer func()
}

func (this *OnlineData) SendMessage(msgId pb.MsgID, data any) {
	body, _ := json.Marshal(data)
	err := this.StreamConn.Send(&pb.GameMessage{MsgId: uint32(msgId), Body: body})
	if err != nil {
		this.MarkOffLine()
	}
}

func (this *OnlineData) MarkOffLine() {

}
