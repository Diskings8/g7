package limiter

import "sync/atomic"

type ConnectionLimiter struct {
	current int64
	max     int64
}

func NewConnectionLimiter(LimitPerSecond int64) *ConnectionLimiter {
	return &ConnectionLimiter{max: LimitPerSecond}
}

func (l *ConnectionLimiter) Allow() bool {
	val := atomic.AddInt64(&l.current, 1)
	if val > l.max {
		atomic.AddInt64(&l.current, -1)
		return false
	}
	return true
}
