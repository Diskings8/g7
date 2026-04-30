package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type GameMonitor struct {
	gameServerPrefix string
	etcdClient       *clientv3.Client
	mu               sync.RWMutex
	cache            map[string]*HashRing
}

func NewGameMonitor() *GameMonitor {
	return &GameMonitor{
		cache:            make(map[string]*HashRing),
		etcdClient:       GetEtcdClient(),
		gameServerPrefix: GetAllGameRpcPrefix(),
	}
}

func (gm *GameMonitor) Run() {
	gm.loadAllGameServers()
	go gm.watchGameServersWithClient()
}

func (gm *GameMonitor) GetRandGameServerAddr(serverID string) (string, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	hashRing, ok := gm.cache[serverID]
	if !ok {
		return "", false
	}
	addr, found := hashRing.GetWorkerByRand()
	return addr, found
}

func (gm *GameMonitor) GetGameServerAddr(serverID, key string) (string, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	hashRing, ok := gm.cache[serverID]
	if !ok {
		return "", false
	}
	addr, found := hashRing.GetWorkerByKey(key)
	return addr, found
}

func (gm *GameMonitor) setGameServerAddr(serverID string, addr string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	// 检查是否已存在相同地址，避免重复
	if _, ok := gm.cache[serverID]; !ok {
		gm.cache[serverID] = NewHashRing()
	}
	if gm.cache[serverID].HasKey(addr) {
		return
	}
	gm.cache[serverID].AddWorker(addr)
}

func (gm *GameMonitor) removeGameServerAddr(serverID string, addr string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.cache[serverID].RemoveWorker(addr)
}

// loadAllGameServers 启动时全量拉取已注册的 GameServer
func (gm *GameMonitor) loadAllGameServers() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 前缀查询：获取 /game_servers/ 下所有节点
	resp, err := gm.etcdClient.Get(ctx, gm.gameServerPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Printf("全量获取游戏服失败: %v", err)
		return
	}

	// 写入本地列表
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)

		// 解析 key：/game_rpc/91001_0/172.0.0.4:8080
		parts := strings.Split(key, "/")
		if len(parts) < 3 {
			continue
		}

		serverID := gm.splitGameServerId(parts[2]) // 提取 91001

		gm.setGameServerAddr(serverID, value)
	}
}

func (gm *GameMonitor) watchGameServersWithClient() {

	// 全量获取档期活跃服务器
	log.Println("开始监听 etcd 游戏服变化：", gm.gameServerPrefix)

	// 监听前缀
	watchChan := gm.etcdClient.Watch(context.Background(), gm.gameServerPrefix, clientv3.WithPrefix())

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

				IdWithInstance := parts[2]
				serverID := gm.splitGameServerId(IdWithInstance) // 提取 91001

				if ev.Type == clientv3.EventTypePut {
					// 新增/更新 → 覆盖（单服只保留一个）
					gm.setGameServerAddr(serverID, value)
					log.Printf("游戏服更新 serverID=%s → %s\n", IdWithInstance, value)
				} else if ev.Type == clientv3.EventTypeDelete {
					// 删除 → 移除
					gm.removeGameServerAddr(serverID, value)
					log.Printf("游戏服下线 serverID=%s\n", IdWithInstance)
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (gm *GameMonitor) splitGameServerId(key string) string {
	parts := strings.Split(key, "_")
	return parts[0]
}
