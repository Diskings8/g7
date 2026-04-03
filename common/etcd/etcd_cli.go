package etcd

import (
	"context"
	"g7/common/globals"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	etcdClient *clientv3.Client
	once       sync.Once
)

// InitETCD 初始化
func InitETCD(Addr string) {
	once.Do(func() {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{Addr},
			DialTimeout: 3 * time.Second,
		})
		if err != nil {
			log.Fatal("etcd 连接失败:", err)
		}
		etcdClient = cli
	})
}

func getGatewayPrefix() string {
	return "/" + globals.GateWays + "/"
}

func getGameServerPrefix(serverID string) string {
	return "/" + globals.GameServer + "/" + serverID + "/"
}

// RegisterGateway 注册网关
func RegisterGateway(addr string) {
	key := getGatewayPrefix() + addr
	registerWithLease(key, addr)
}

// RegisterGameServer 注册游戏服
func RegisterGameServer(serverID string, addr string) {
	key := getGameServerPrefix(serverID) + addr
	registerWithLease(key, addr)
}

// 内部通用：带租约注册
func registerWithLease(key, value string) {
	resp, err := etcdClient.Grant(context.Background(), 10)
	if err != nil {
		log.Printf("etcd租约失败: %v", err)
		return
	}
	_, err = etcdClient.Put(context.Background(), key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		log.Printf("etcd注册失败: %v", err)
		return
	}
	// 自动续租
	ch, err := etcdClient.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return
	}
	go func() {
		for range ch {
		}
	}()
}

// GetAllGateways 获取所有网关
func GetAllGateways() ([]string, error) {
	resp, err := etcdClient.Get(context.Background(), getGatewayPrefix(), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var addrs []string
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}
	return addrs, nil
}

// GetGameServersByServerID 获取指定区服的游戏服
func GetGameServersByServerID(serverID string) ([]string, error) {
	key := getGameServerPrefix(serverID)
	resp, err := etcdClient.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var addrs []string
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}
	return addrs, nil
}
