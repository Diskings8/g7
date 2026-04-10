package tcp_session

import (
	"g7/common/protos/pb"
	"net"
	"sync"
)

// Session 会话：网关只存这些！绝对不存业务数据！
type Session struct {
	conn       net.Conn
	userID     int64 // 用户ID
	playerID   int64 // 角色ID
	serverID   int32 // 要连接的游戏服ID
	gameStream pb.GameStreamService_StreamClient
	closed     bool
	lock       sync.Mutex
}

var (
	sessionMap = make(map[net.Conn]*Session)
	sessLock   sync.RWMutex
)

func NewSession(conn net.Conn) *Session {
	sess := &Session{conn: conn}
	sessLock.Lock()
	sessionMap[conn] = sess
	sessLock.Unlock()
	return sess
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

	sessLock.Lock()
	delete(sessionMap, s.conn)
	sessLock.Unlock()
}

func (s *Session) SetOwner(UerId, PlayerId int64, serverId int32) {
	s.userID = UerId
	s.playerID = PlayerId
	s.serverID = serverId
}

func (s *Session) SetStream(Stream pb.GameStreamService_StreamClient) {
	s.gameStream = Stream
}

func (s *Session) GetPlayerId() int64 {
	return s.playerID
}
