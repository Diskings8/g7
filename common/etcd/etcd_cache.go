package etcd

import (
	"context"
	"g7/common/globals"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"sync"
	"time"
)

// 全局服务缓存（生产级）
var (
	GameServerCache = &serviceCache{
		cache: make(map[string]string), // key: serverID, value: addr
	}
	GatewayCache = &serviceCache{
		cache: make(map[string]string),
	}
)

// 真正的缓存结构（带锁）
type serviceCache struct {
	mu    sync.RWMutex
	cache map[string]string // 单服模式：1个serverID只存1个地址
}

func GetGameServerAddr(serverID string) (string, bool) {
	GameServerCache.mu.RLock()
	defer GameServerCache.mu.RUnlock()
	addr, ok := GameServerCache.cache[serverID]
	return addr, ok
}

func WatchGameServers() {
	WatchGameServersWithClient(etcdClient)
}

// WatchGameServersWithClient 变化（生产级核心）
func WatchGameServersWithClient(etcdClient *clientv3.Client) {
	prefix := "/" + globals.GameServer + "/"
	log.Println("开始监听 etcd 游戏服变化：", prefix)

	// 监听前缀
	watchChan := etcdClient.Watch(context.Background(), prefix, clientv3.WithPrefix())

	for {
		for resp := range watchChan {
			for _, ev := range resp.Events {
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)

				// 解析 key：/game_server/91001/127.0.0.1:8082
				parts := strings.Split(key, "/")
				if len(parts) < 3 {
					continue
				}

				serverID := parts[2] // 提取 91001

				GameServerCache.mu.Lock()
				if ev.Type == clientv3.EventTypePut {
					// 新增/更新 → 覆盖（单服只保留一个）
					GameServerCache.cache[serverID] = value
					log.Printf("游戏服更新 serverID=%s → %s\n", serverID, value)
				} else if ev.Type == clientv3.EventTypeDelete {
					// 删除 → 移除
					delete(GameServerCache.cache, serverID)
					log.Printf("游戏服下线 serverID=%s\n", serverID)
				}
				GameServerCache.mu.Unlock()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
