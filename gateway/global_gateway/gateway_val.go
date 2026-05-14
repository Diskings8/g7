package global_gateway

import (
	"g7/gateway/conn_session"
	"sync/atomic"
)

var GCurrentConnection atomic.Int32

func GetConnCount() int32 {
	return GCurrentConnection.Load()
}

var GConnSessionMap = conn_session.GetSessionManager()

func AllSessionMapInit() {
	GConnSessionMap.Init()
}
