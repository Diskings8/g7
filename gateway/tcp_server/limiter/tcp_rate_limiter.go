package limiter

import (
	"sync/atomic"
	"time"
)

type RateLimiter struct {
	lastResetTime       int64
	gatewayCurrentCount int32
	gatewayRateLimit    int32
}

func NewRateLimiter(LimitPerSecond int32) *RateLimiter {
	return &RateLimiter{gatewayRateLimit: LimitPerSecond}
}

func (l *RateLimiter) Allow() bool {
	now := time.Now().Unix()
	if now != atomic.LoadInt64(&l.lastResetTime) {
		atomic.StoreInt32(&l.gatewayCurrentCount, 0)
		atomic.StoreInt64(&l.lastResetTime, now)
	}
	return atomic.AddInt32(&l.gatewayCurrentCount, 1) < l.gatewayRateLimit
}
