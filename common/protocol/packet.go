package protocol

import (
	"encoding/binary"
	"g7/common/protos/pb"
	"io"
	"net"
	"sync"
)

var msgBufPool = sync.Pool{
	New: func() interface{} {
		// 这里按你的业务设置一个合理的初始大小，比如 4KB
		// 实际会根据需要扩容，但池里只会保存固定大小的对象
		return make([]byte, 4096)
	},
}

var headerBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, HeaderSize)
	},
}

// ReadMessage 从TCP连接读取一条完整消息（解决粘包）
func ReadMessage(conn net.Conn) (*Message, error) {
	headerBuf := getHeadBuf(HeaderSize)
	defer putHeadBuf(headerBuf) // 读取完归还
	if _, err := io.ReadFull(conn, headerBuf); err != nil {
		return nil, err
	}

	// 解析长度和协议ID
	length := binary.BigEndian.Uint32(headerBuf[:headSizeIndex])
	msgID := binary.BigEndian.Uint16(headerBuf[headSizeIndex:msgSizeIndex])

	// 读取body
	bodyBuf := getMsgBuf(int(length))
	defer putMsgBuf(bodyBuf)

	if _, err := io.ReadFull(conn, bodyBuf); err != nil {
		return nil, err
	}

	return &Message{
		Length: length,
		MsgID:  pb.MsgID(msgID),
		Body:   bodyBuf,
	}, nil
}

// WriteMessage 发送消息
func WriteMessage(conn net.Conn, msgID pb.MsgID, body []byte) error {
	length := uint32(len(body))
	totalSize := HeaderSize + len(body)

	// 从池里获取缓冲区
	buf := getMsgBuf(totalSize)
	defer putMsgBuf(buf) // 发送完归还

	binary.BigEndian.PutUint32(buf[:headSizeIndex], length)
	binary.BigEndian.PutUint16(buf[headSizeIndex:msgSizeIndex], uint16(msgID))
	copy(buf[msgSizeIndex:], body)

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
