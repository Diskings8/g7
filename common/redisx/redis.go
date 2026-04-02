package redisx

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client
var ctx = context.Background()

func Init(addr, password string, db int) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		panic("redis 连接失败: " + err.Error())
	}
}
