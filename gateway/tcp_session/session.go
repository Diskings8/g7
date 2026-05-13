package tcp_session

import (
	"fmt"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

// Session 会话：网关只存这些！绝对不存业务数据！
type Session struct {
	conn       net.Conn
	userID     int64 // 用户ID
	playerID   int64 // 角色ID
	serverID   int32 // 要连接的游戏服ID
	seq        uint32
	gameStream pb.GameStreamService_StreamClient
	roomStream pb.RoomStreamService_StreamClient
	roomId     string
	closed     bool
	lock       sync.Mutex
	connSend   chan *pb.GameMessage
	gameSend   chan *pb.GameMessage
	roomSend   chan *pb.GameMessage
	// 限流
	pktCount            int32
	pktTime             int64
	etcdGatewayGrpcAddr string
}

func NewSession(conn net.Conn, etcdGatewayGrpcAddr string) *Session {
	sess := &Session{conn: conn}
	sess.Init(etcdGatewayGrpcAddr)
	gSessionManager.NewSession(sess, conn)
	return sess
}

func (s *Session) Init(etcdGatewayGrpcAddr string) {
	s.connSend = make(chan *pb.GameMessage, 400)
	s.roomSend = make(chan *pb.GameMessage, 400)
	s.gameSend = make(chan *pb.GameMessage, 400)
	s.etcdGatewayGrpcAddr = etcdGatewayGrpcAddr
}

func (s *Session) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	s.conn.Close()
	if s.gameStream != nil {
		_ = s.gameStream.CloseSend()
	}

	gSessionManager.RemoveConn(s.conn)

	close(s.connSend)
	close(s.gameSend)
	close(s.roomSend)
}

func (s *Session) SetOwner(UerId, PlayerId int64, serverId int32) {
	s.userID = UerId
	s.playerID = PlayerId
	s.serverID = serverId
}

func (s *Session) newSeq() uint32 {
	s.seq++
	return s.seq
}

func (s *Session) SetGameStream(Stream pb.GameStreamService_StreamClient) {
	s.gameStream = Stream
}

func (s *Session) SetRoomStream(Stream pb.RoomStreamService_StreamClient, roomId string) {
	s.roomStream = Stream
	s.roomId = roomId
}

func (s *Session) GetPlayerId() int64 {
	return s.playerID
}

func (s *Session) GetServerId() int32 {
	return s.serverID
}

func (s *Session) AllowPacket() bool {
	now := time.Now().Unix()
	if atomic.LoadInt64(&s.pktTime) != now {
		atomic.StoreInt32(&s.pktCount, 1)
		atomic.StoreInt64(&s.pktTime, now)
		return true
	}
	return atomic.AddInt32(&s.pktCount, 1) <= 50 // 每秒50包，防止攻击
}

func (s *Session) sendToConn(msg *pb.GameMessage) {
	if s.connSend != nil {
		s.connSend <- msg
	}
}

func (s *Session) AuthToGame() {
	msg := pb.Req_AuthClientToGame{PlayerID: s.GetPlayerId(), ServerID: s.GetServerId(), GatewayAddr: s.etcdGatewayGrpcAddr}
	msgBody, _ := proto.Marshal(&msg)
	req := &pb.GameMessage{MsgId: uint32(pb.MsgID_MSG_AUTH), Body: msgBody}
	err := s.gameStream.Send(req)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func (s *Session) AuthToRoom() {
	msg := pb.Req_AuthClientToRoom{PlayerID: s.GetPlayerId(), ServerID: s.GetServerId(), RoomId: s.roomId}
	msgBody, _ := proto.Marshal(&msg)
	req := &pb.GameMessage{MsgId: uint32(pb.MsgID_MSG_AUTH), Body: msgBody}
	err := s.roomStream.Send(req)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func (s *Session) RunGoRoutineToRecvFromConn() {
	for {
		packet, err := protocol.ReadPacketFromConn(s.conn)
		if err != nil {
			log.Printf("客户端断开: %v", err)
			return
		}
		if packet.MsgID == pb.MsgID_MSG_AUTH {
			logger.Log.Info("再次认证")
			continue
		}
		gameMessage := &pb.GameMessage{MsgId: uint32(packet.MsgID), Body: packet.Body}
		s.switchToSelectStream(gameMessage)
		packet.Release()
	}
}

func (s *Session) switchToSelectStream(gameMessage *pb.GameMessage) {
	msgId := pb.MsgID(gameMessage.GetMsgId())
	if msgId >= pb.MsgID_MSG_GATEWAY_TO_SCENE && msgId < pb.MsgID_MSG_GATEWAY_TO_GAME {
		s.roomSend <- gameMessage
		return
	}
	s.gameSend <- gameMessage
	return
}

func (s *Session) RunGoRoutineToSendToConn() {
	for !s.closed {
		select {
		case msg, ok := <-s.connSend:
			if !ok {
				return
			}
			err := protocol.WritePacketToConn(s.conn, pb.MsgID(msg.MsgId), s.newSeq(), msg.Body)
			if err != nil {
				log.Printf("write to conn error: %v", err)
				s.Close()
				return
			}
		}
	}
}

// RunGoRoutineToRecvFromGame 游戏服 → 网关 → 客户端
func (s *Session) RunGoRoutineToRecvFromGame() {
	defer func() {
		if e := recover(); e != nil {
			logger.Log.Error(fmt.Sprintf("%v", e))
		}
	}()
	for !s.closed {
		pkt, err := s.gameStream.Recv()
		if err != nil {
			log.Printf("游戏服流断开: %v", err)
			s.Close()
			return
		}
		s.sendToConn(pkt)
	}
}

// RunGoRoutineToSendToGame 客户端 → 网关 → 游戏服
func (s *Session) RunGoRoutineToSendToGame() {
	defer func() {
		if e := recover(); e != nil {
			logger.Log.Error(fmt.Sprintf("%v", e))
		}
	}()
	s.AuthToGame()
	for !s.closed {
		select {
		case msg, ok := <-s.gameSend:
			if !ok {
				return
			}
			_ = s.gameStream.Send(&pb.GameMessage{
				MsgId: msg.MsgId,
				Body:  msg.Body,
			})
		}
	}
}

func (s *Session) RunGoRoutineToSendToRoom() {
	defer func() {
		if e := recover(); e != nil {
			logger.Log.Error(fmt.Sprintf("%v", e))
		}
	}()
	s.AuthToRoom()
	for !s.closed {
		select {
		case msg, ok := <-s.roomSend:
			if !ok {
				return
			}
			_ = s.roomStream.Send(&pb.GameMessage{
				MsgId: msg.MsgId,
				Body:  msg.Body,
			})
		}
	}
}

func (s *Session) RunGoRoutineToRecvFromRoom() {
	defer func() {
		if e := recover(); e != nil {
			logger.Log.Error(fmt.Sprintf("%v", e))
		}
	}()
	for !s.closed {
		pkt, err := s.roomStream.Recv()
		if err != nil {
			log.Printf("%d 房间服流断开: %v", s.playerID, err)
			return
		}
		s.sendToConn(pkt)
	}
}
