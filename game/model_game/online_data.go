package model_game

import (
	"fmt"
	"g7/common/logger"
	"g7/common/protos/pb"
	"github.com/golang/protobuf/proto"
	"time"
)

type OnlineData struct {
	// 流相关
	StreamId         uint64
	StreamConn       pb.GameStreamService_StreamServer
	StreamCancelFunc func()              // 断开流
	IsChanClosed     bool                // 标记是否关闭
	LastHearBeatTime time.Time           // 上次心跳时间
	ActionChan       chan func()         // 独立协程队列（所有逻辑都走这里）
	QuitChan         chan struct{}       // 退出用
	SaveChan         chan SaveDaoD       // 给存储系统用的channel
	RpcSendChan      chan pb.GameMessage // 发送协议

	// limit
	LimitLastReqTime int64
	LimitReqCount    int32
}

func (this *OnlineData) SendMessage(msgId pb.MsgID, data proto.Message) {
	body, err := proto.Marshal(data)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("%+v: %s", data, err.Error()))
		return
	}
	this.RpcSendChan <- pb.GameMessage{MsgId: uint32(msgId), Body: body}
}

func (this *OnlineData) GetLastHearBeatTime() time.Time {
	return this.LastHearBeatTime
}

func (this *OnlineData) Init(StreamConn pb.GameStreamService_StreamServer, systemSaveChannel chan SaveDaoD) {
	this.StreamConn = StreamConn
	this.IsChanClosed = false
	this.ActionChan = make(chan func(), 1000)
	this.RpcSendChan = make(chan pb.GameMessage, 1000)
	this.QuitChan = make(chan struct{}, 1)
	this.SaveChan = systemSaveChannel
	this.LastHearBeatTime = time.Now()
}
