package limiter

import (
	"sync"
	"time"
)

type IPLimiter struct {
	counts    map[string]int32
	expire    map[string]int64
	mu        sync.RWMutex
	maxPerSec int32 // 每秒最大请求数
}

func NewIPLimiter(LimitPerSecond int32) *IPLimiter {
	return &IPLimiter{
		counts:    make(map[string]int32),
		expire:    make(map[string]int64),
		maxPerSec: LimitPerSecond,
	}
}

func (l *IPLimiter) Allow(ip string) bool {
	now := time.Now().Unix()
	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果过期了，重置计数
	if now > l.expire[ip] {
		l.counts[ip] = 1
		l.expire[ip] = now + 1
		return true
	}

	l.counts[ip]++
	return l.counts[ip] < l.maxPerSec
}
