package tcp_session

import (
	"net"
	"sync"
)

var gSessionManager = &SessionManager{}

type SessionManager struct {
	sessionMap map[net.Conn]*Session
	rw         sync.RWMutex
}

func GetSessionManager() *SessionManager {
	return gSessionManager
}

func (sm *SessionManager) Init() {
	sm.sessionMap = make(map[net.Conn]*Session)
}

func (sm *SessionManager) NewSession(session *Session, conn net.Conn) {
	sm.rw.Lock()
	defer sm.rw.Unlock()
	sm.sessionMap[conn] = session
}

func (sm *SessionManager) RemoveConn(conn net.Conn) {
	sm.rw.Lock()
	defer sm.rw.Unlock()
	delete(sm.sessionMap, conn)
}

func (sm *SessionManager) FindSessionByPlayerId(playerId int64) *Session {
	sm.rw.RLock()
	defer sm.rw.RUnlock()
	for _, v := range sm.sessionMap {
		if v.playerID == playerId {
			return v
		}
	}
	return nil
}
