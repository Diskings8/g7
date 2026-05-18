package conn_session

import (
	"g7/common/netc"
	"sync"
)

var gSessionManager = &SessionManager{}

type SessionManager struct {
	sessionMap map[netc.NetConnInterface]*Session
	rw         sync.RWMutex
}

func GetSessionManager() *SessionManager {
	return gSessionManager
}

func (sm *SessionManager) Init() {
	sm.sessionMap = make(map[netc.NetConnInterface]*Session)
}

func (sm *SessionManager) NewSession(conn netc.NetConnInterface, session *Session) {
	sm.rw.Lock()
	defer sm.rw.Unlock()
	sm.sessionMap[conn] = session
}

func (sm *SessionManager) RemoveConn(conn netc.NetConnInterface) {
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

func (sm *SessionManager) AllClose() {
	sm.rw.RLock()
	defer sm.rw.RUnlock()
	for _, v := range sm.sessionMap {
		v.Close()
	}
}
