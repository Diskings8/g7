package global_gateway

import "sync/atomic"

var GCurrentConnection atomic.Int32

func GetConnCount() int32 {
	return GCurrentConnection.Load()
}
