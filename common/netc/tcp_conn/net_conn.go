package tcp_conn

import (
	"encoding/binary"
	"g7/common/netc"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"io"
	"net"
)

type NetConn struct {
	conn net.Conn
}

var _ netc.NetConnInterface = (*NetConn)(nil)

func NewNetConn(c net.Conn) *NetConn {
	return &NetConn{conn: c}
}

func (nc *NetConn) Close() error {
	return nc.conn.Close()
}

func (nc *NetConn) ReadFromConn() (*protocol.Packet, error) {
	return readPacketFromConn(nc.conn)
}

func (nc *NetConn) WriteToConn(seq uint32, message *pb.GameMessage) error {
	return writePacketToConn(nc.conn, pb.MsgID(message.MsgId), seq, message.Body)
}

func (nc *NetConn) RemoteAddr() string {
	return nc.conn.RemoteAddr().String()
}

func readPacketFromConn(conn net.Conn) (*protocol.Packet, error) {
	headerBuf := protocol.GetHeadBuf(protocol.TcpHeaderSize)
	defer protocol.PutHeadBuf(headerBuf) // 读取完归还
	if _, err := io.ReadFull(conn, headerBuf); err != nil {
		return nil, err
	}

	// 解析长度和协议ID
	length := binary.BigEndian.Uint32(headerBuf[0:protocol.TcpLengthSizeTail])
	seq := binary.BigEndian.Uint32(headerBuf[protocol.TcpLengthSizeTail:protocol.TcpSeqSizeTail])
	msgID := binary.BigEndian.Uint32(headerBuf[protocol.TcpSeqSizeTail:protocol.TcpMsgIdSizeTail])

	// 读取body
	bodyBuf := protocol.GetMsgBuf(int(length - protocol.TcpHeaderSize))

	if _, err := io.ReadFull(conn, bodyBuf); err != nil {
		protocol.PutMsgBuf(bodyBuf)
		return nil, err
	}

	return &protocol.Packet{
		Length: length,
		Seq:    seq,
		MsgID:  pb.MsgID(msgID),
		Body:   bodyBuf,
	}, nil
}

func writePacketToConn(conn net.Conn, msgID pb.MsgID, seq uint32, body []byte) error {
	totalSize := protocol.TcpHeaderSize + len(body)

	// 从池里获取缓冲区
	buf := protocol.GetMsgBuf(totalSize)
	defer protocol.PutMsgBuf(buf) // 发送完归还

	binary.BigEndian.PutUint32(buf[:protocol.TcpLengthSizeTail], uint32(totalSize))
	binary.BigEndian.PutUint32(buf[protocol.TcpLengthSizeTail:protocol.TcpSeqSizeTail], seq)
	binary.BigEndian.PutUint32(buf[protocol.TcpSeqSizeTail:protocol.TcpMsgIdSizeTail], uint32(msgID))
	copy(buf[protocol.TcpMsgIdSizeTail:], body)

	_, err := conn.Write(buf)
	return err
}
