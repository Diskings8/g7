package global_gateway

import (
	"g7/gateway/tcp_session"
	"sync/atomic"
)

var GCurrentConnection atomic.Int32

func GetConnCount() int32 {
	return GCurrentConnection.Load()
}

var GConnSessionMap = tcp_session.GetSessionManager()
