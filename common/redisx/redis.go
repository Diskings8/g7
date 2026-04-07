package redisx

import (
	"context"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"time"
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

func GetKey(key string) (string, error) {
	return RDB.Get(context.Background(), key).Result()
}

func SetKey(key string, value []byte, cacheExpire time.Duration) error {
	return RDB.Set(context.Background(), key, value, cacheExpire+time.Duration(rand.Intn(120))*time.Minute).Err()
}
