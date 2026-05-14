package websocket_conn

import (
	"encoding/binary"
	"errors"
	"g7/common/netc"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"github.com/gorilla/websocket"
	"log"
)

type WBConn struct {
	conn *websocket.Conn
}

var _ netc.NetConnInterface = (*WBConn)(nil)

func NewWebSocketConn(c *websocket.Conn) *WBConn {
	return &WBConn{conn: c}
}

func (wbc *WBConn) Close() error {
	return wbc.conn.Close()
}

func (wbc *WBConn) ReadFromConn() (*protocol.Packet, error) {
	mt, msg, err := wbc.conn.ReadMessage()
	switch mt {
	case websocket.BinaryMessage:
		MsgId := binary.BigEndian.Uint32(msg[0:protocol.WBMsgIdSize])
		MsgBody := msg[protocol.WBMsgIdSize:]
		result := protocol.Packet{
			Length: uint32(len(msg)),
			Seq:    0,
			MsgID:  pb.MsgID(MsgId),
			Body:   MsgBody,
		}
		return &result, nil
	case websocket.PingMessage:
		result := protocol.Packet{
			Length: uint32(len(msg)),
			Seq:    0,
			MsgID:  pb.MsgID_MSG_HeartBeat,
			Body:   nil,
		}
		return &result, nil
	default:
		log.Println(mt, msg, err)
		if err == nil {
			return nil, errors.New("other type")
		}
		return nil, err
	}
}

func (wbc *WBConn) WriteToConn(seq uint32, message *pb.GameMessage) error {
	totalSize := protocol.WBMsgIdSize + len(message.GetBody())
	buf := protocol.GetMsgBuf(totalSize)
	defer protocol.PutMsgBuf(buf) // 发送完归还
	binary.BigEndian.PutUint32(buf[0:protocol.WBMsgIdSize], message.GetMsgId())
	copy(buf[protocol.WBMsgIdSize:], message.GetBody())
	err := wbc.conn.WriteMessage(websocket.BinaryMessage, buf)
	return err
}

func (wbc *WBConn) RemoteAddr() string {
	return wbc.conn.RemoteAddr().String()
}
