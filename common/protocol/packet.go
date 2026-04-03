package protocol

import (
	"context"
	"encoding/binary"
	"g7/common/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
)

// ReadMessage 从TCP连接读取一条完整消息（解决粘包）
func ReadMessage(conn net.Conn) (*Message, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	// 解析长度和协议ID
	length := binary.BigEndian.Uint32(header[:headSizeIndex])
	msgID := binary.BigEndian.Uint32(header[headSizeIndex:msgSizeIndex])

	// 读取body
	body := make([]byte, length)
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, err
	}

	return &Message{
		Length: length,
		MsgID:  pb.MsgID(msgID),
		Body:   body,
	}, nil
}

// WriteMessage 发送消息
func WriteMessage(conn net.Conn, msgID pb.MsgID, body []byte) error {
	length := uint32(len(body))
	buf := make([]byte, HeaderSize+len(body))

	binary.BigEndian.PutUint32(buf[:headSizeIndex], length)
	binary.BigEndian.PutUint32(buf[headSizeIndex:msgSizeIndex], uint32(msgID))
	copy(buf[msgSizeIndex:], body)

	_, err := conn.Write(buf)
	return err
}

func NewGameNodeClient(ctx context.Context, addr string) (pb.GameNodeServiceClient, error) {
	// 1. 建立连接（单工不需要长流）
	conn, err := grpc.DialContext(ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	// 2. 创建单工客户端
	client := pb.NewGameNodeServiceClient(conn)
	return client, nil
}
