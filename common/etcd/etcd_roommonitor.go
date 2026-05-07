package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"sync"
	"time"
)

type ComMonitor struct {
	serverPrefix string
	etcdClient   *clientv3.Client
	mu           sync.RWMutex
	cache        map[string]string
}

func NewComMonitor(Prefix string) *ComMonitor {
	return &ComMonitor{
		cache:        make(map[string]string),
		etcdClient:   GetEtcdClient(),
		serverPrefix: Prefix,
	}
}

func (cm *ComMonitor) Run() {
	cm.loadAllServers()
	go cm.watchServersWithClient()
}

func (cm *ComMonitor) GetRandServerAddr() (string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if len(cm.cache) == 0 {
		return "", false
	}
	for _, v := range cm.cache {
		return v, true
	}
	return "", false
}

func (cm *ComMonitor) GetServerAddr(key string) (string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	addr, ok := cm.cache[key]
	return addr, ok
}

func (cm *ComMonitor) setGameServerAddr(key string, addr string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cache[key] = addr
}

func (cm *ComMonitor) removeServerAddr(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.cache, key)
}

func (cm *ComMonitor) loadAllServers() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 前缀查询：获取 /game_servers/ 下所有节点
	resp, err := cm.etcdClient.Get(ctx, cm.serverPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Printf("全量获取失败: %v", err)
		return
	}

	// 写入本地列表
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)

		// 解析 key：/game_rpc/91001_0/172.0.0.4:8080
		parts := strings.Split(key, "/")
		if len(parts) < 2 {
			continue
		}

		cm.setGameServerAddr(parts[2], value)
	}
}

func (cm *ComMonitor) watchServersWithClient() {
	// 监听前缀
	watchChan := cm.etcdClient.Watch(context.Background(), cm.serverPrefix, clientv3.WithPrefix())

	for {
		for resp := range watchChan {
			for _, ev := range resp.Events {
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)

				// 解析 key：/game_server/91001/127.0.0.1:8082
				parts := strings.Split(key, "/")
				if len(parts) < 2 {
					continue
				}

				if ev.Type == clientv3.EventTypePut {
					// 新增/更新 → 覆盖（单服只保留一个）
					cm.setGameServerAddr(parts[2], value)
				} else if ev.Type == clientv3.EventTypeDelete {
					// 删除 → 移除
					cm.removeServerAddr(parts[2])
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
