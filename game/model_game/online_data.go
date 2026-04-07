package model_game

import (
	"encoding/json"
	"g7/common/protos/pb"
	"g7/game/const_game"
	"sync/atomic"
	"time"
)

type OnlineData struct {
	StreamConn            pb.GameStreamService_StreamServer `gorm:"-"`
	IsChanClosed          bool                              `gorm:"-"` // 标记是否关闭
	saveTicker            *time.Ticker                      // 玩家自己的定时器
	IsDisconnecting       *atomic.Bool
	DisconnectCancelTimer func()        `gorm:"-"`
	ActionChan            chan func()   `gorm:"-"` // 独立协程队列（所有逻辑都走这里）
	QuitChan              chan struct{} // 退出用
	SaveChan              chan *PlayerDao
}

func (this *OnlineData) SendMessage(msgId pb.MsgID, data any) {
	body, _ := json.Marshal(data)
	_ = this.StreamConn.Send(&pb.GameMessage{MsgId: uint32(msgId), Body: body})
}

func (this *OnlineData) Init(StreamConn pb.GameStreamService_StreamServer, c chan *PlayerDao) {
	this.StreamConn = StreamConn
	this.IsChanClosed = false
	this.saveTicker = time.NewTicker(const_game.SaveDaoTimeTick)
	this.IsDisconnecting = &atomic.Bool{}
	this.DisconnectCancelTimer = nil
	this.ActionChan = make(chan func(), 1000)
	this.QuitChan = make(chan struct{}, 1)
	this.SaveChan = c
}
