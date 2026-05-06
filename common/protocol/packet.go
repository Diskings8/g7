package protocol

import (
	"encoding/binary"
	"g7/common/protos/pb"
	"io"
	"net"
)

type Packet struct {
	Length uint32
	Seq    uint32
	MsgID  pb.MsgID
	Body   []byte
}

func (p *Packet) Release() {
	msgBufPool.Put(p.Body)
	p.Body = nil
}

// ReadPacketFromConn 从TCP连接读取一条完整消息（解决粘包）
func ReadPacketFromConn(conn net.Conn) (*Packet, error) {
	headerBuf := getHeadBuf(HeaderSize)
	defer putHeadBuf(headerBuf) // 读取完归还
	if _, err := io.ReadFull(conn, headerBuf); err != nil {
		return nil, err
	}

	// 解析长度和协议ID
	length := binary.BigEndian.Uint32(headerBuf[0:lengthSizeTail])
	seq := binary.BigEndian.Uint32(headerBuf[lengthSizeTail:seqSizeTail])
	msgID := binary.BigEndian.Uint32(headerBuf[seqSizeTail:msgIdSizeTail])

	// 读取body
	bodyBuf := getMsgBuf(int(length - HeaderSize))

	if _, err := io.ReadFull(conn, bodyBuf); err != nil {
		putMsgBuf(bodyBuf)
		return nil, err
	}

	return &Packet{
		Length: length,
		Seq:    seq,
		MsgID:  pb.MsgID(msgID),
		Body:   bodyBuf,
	}, nil
}

// WritePacketToConn 发送消息
func WritePacketToConn(conn net.Conn, msgID pb.MsgID, seq uint32, body []byte) error {
	totalSize := HeaderSize + len(body)

	// 从池里获取缓冲区
	buf := getMsgBuf(totalSize)
	defer putMsgBuf(buf) // 发送完归还

	binary.BigEndian.PutUint32(buf[:lengthSizeTail], uint32(totalSize))
	binary.BigEndian.PutUint32(buf[lengthSizeTail:seqSizeTail], seq)
	binary.BigEndian.PutUint32(buf[seqSizeTail:msgIdSizeTail], uint32(msgID))
	copy(buf[msgIdSizeTail:], body)

	_, err := conn.Write(buf)
	return err
}

func getMsgBuf(size int) []byte {
	buf := msgBufPool.Get().([]byte)
	if cap(buf) < size {
		// 如果容量不够，扩容
		buf = make([]byte, size)
	} else {
		// 裁剪到需要的长度，避免写入时超出
		buf = buf[:size]
	}
	return buf
}

func putMsgBuf(buf []byte) {
	msgBufPool.Put(buf[:cap(buf)]) // 归还完整容量的slice
}

func getHeadBuf(size int) []byte {
	buf := headerBufPool.Get().([]byte)
	if cap(buf) < size {
		// 如果容量不够，扩容
		buf = make([]byte, size)
	} else {
		// 裁剪到需要的长度，避免写入时超出
		buf = buf[:size]
	}
	return buf
}

func putHeadBuf(buf []byte) {
	// 重置缓冲区，避免脏数据
	for i := range buf {
		buf[i] = 0
	}
	headerBufPool.Put(buf[:cap(buf)]) // 归还完整容量的slice
}
