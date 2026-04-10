package redisx

import (
	"sync/atomic"
	"time"
)

type RedisBreaker struct {
	failCount int32
	breakTime int64
}

func (r *RedisBreaker) Allow() bool {
	now := time.Now().Unix()
	if now < atomic.LoadInt64(&r.breakTime) {
		return false // 熔断中
	}
	return true
}

func (r *RedisBreaker) Fail() {
	if atomic.AddInt32(&r.failCount, 1) >= 3 {
		atomic.StoreInt64(&r.breakTime, time.Now().Unix()+5)
	}
}

func (r *RedisBreaker) Success() {
	atomic.StoreInt32(&r.failCount, 0)
}
