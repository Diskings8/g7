package protocol

import "g7/common/protos/pb"

// 游戏协议头（固定 6 字节）
// [4字节长度][2字节协议ID]

const (
	HeaderSize    = 4 + 2 // 长度固定
	headSizeIndex = 4
	msgSizeIndex  = 6
)

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
