package protocol

import "g7/common/protos/pb"

// Message 消息结构
type Message struct {
	Length uint32
	MsgID  pb.MsgID
	Body   []byte
}

func (m *Message) Release() {
	msgBufPool.Put(m.Body)
	m.Body = nil
}
