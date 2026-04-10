package redisx

import (
	"context"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var rDB *redis.Client
var ctx = context.Background()

func Init(addr, password string, db int) {
	rDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := rDB.Ping(ctx).Result()
	if err != nil {
		panic("redis 连接失败: " + err.Error())
	}
}

func GetKey(key string) (string, error) {
	return rDB.Get(context.Background(), key).Result()
}

func SetKey(key string, value []byte, cacheExpire time.Duration) error {
	return rDB.Set(context.Background(), key, value, cacheExpire+time.Duration(rand.Intn(120))*time.Minute).Err()
}

func GetUsedMemoryMB() int {
	// 执行 INFO memory 命令
	info, err := rDB.Info(context.Background(), "memory").Result()
	if err != nil {
		return 0
	}

	// 按行解析
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		// 找到 used_memory 字段
		if strings.HasPrefix(line, "used_memory:") {
			kv := strings.Split(line, ":")
			if len(kv) != 2 {
				return 0
			}

			// 转数字
			usedBytes, _ := strconv.ParseInt(kv[1], 10, 64)
			return int(usedBytes / 1024 / 1024) // 转 MB
		}
	}

	return 0
}
