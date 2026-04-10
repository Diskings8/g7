package tcp_server

import (
	"context"
	"g7/common/etcd"
	"g7/common/globals"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"sync"
	"time"
)

var GTServer = &GatewayTcpServer{}

type GatewayTcpServer struct {
	gameServerPrefix string
	etcdClient       *clientv3.Client
	mu               sync.RWMutex
	cache            map[string]string // 单服模式：1个serverID只存1个地址
}

func (gts *GatewayTcpServer) Init() {
	gts.gameServerPrefix = "/" + globals.GameServer + "/"
	gts.etcdClient = etcd.GetEtcdClient()
	gts.cache = make(map[string]string)

	gts.loadAllGameServers()
	gts.watchGameServersWithClient()
}

func (gts *GatewayTcpServer) getGameServerAddr(serverID string) (string, bool) {
	gts.mu.RLock()
	defer gts.mu.RUnlock()
	addr, ok := gts.cache[serverID]
	return addr, ok
}

func (gts *GatewayTcpServer) setGameServerAddr(serverID string, addr string) {
	gts.mu.Lock()
	defer gts.mu.Unlock()
	gts.cache[serverID] = addr
}

// loadAllGameServers 启动时全量拉取已注册的 GameServer
func (gts *GatewayTcpServer) loadAllGameServers() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 前缀查询：获取 /game_servers/ 下所有节点
	resp, err := gts.etcdClient.Get(ctx, gts.gameServerPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Printf("全量获取游戏服失败: %v", err)
		return
	}

	// 写入本地列表
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)

		// 解析 key：/game_server/91001/127.0.0.1:8082
		parts := strings.Split(key, "/")
		if len(parts) < 3 {
			continue
		}

		serverID := parts[2] // 提取 91001

		gts.mu.Lock()
		gts.cache[serverID] = value
		gts.mu.Unlock()
	}
}

func (gts *GatewayTcpServer) watchGameServersWithClient() {

	// 全量获取档期活跃服务器
	log.Println("开始监听 etcd 游戏服变化：", gts.gameServerPrefix)

	// 监听前缀
	watchChan := gts.etcdClient.Watch(context.Background(), gts.gameServerPrefix, clientv3.WithPrefix())

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

				gts.mu.Lock()
				if ev.Type == clientv3.EventTypePut {
					// 新增/更新 → 覆盖（单服只保留一个）
					gts.cache[serverID] = value
					log.Printf("游戏服更新 serverID=%s → %s\n", serverID, value)
				} else if ev.Type == clientv3.EventTypeDelete {
					// 删除 → 移除
					delete(gts.cache, serverID)
					log.Printf("游戏服下线 serverID=%s\n", serverID)
				}
				gts.mu.Unlock()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
