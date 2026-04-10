package etcd

import (
	"context"
	"fmt"
	"g7/common/globals"
	"log"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
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

func GetEtcdClient() *clientv3.Client {
	return etcdClient
}

func getGatewayPrefix() string {
	return "/" + globals.GateWays + "/"
}

func getGatewayServerPrefix() string {
	return "/" + globals.GateWayServer + "/"
}

func getGameServerPrefix(serverID string) string {
	return "/" + globals.GameServer + "/" + serverID + "/"
}
func getLoginPrefix() string {
	return "/" + globals.LoginServer + "/"
}

// RegisterGateway 注册网关
func RegisterGateway(addr string) {
	key := getGatewayPrefix() + addr
	registerWithLease(key, addr)
}

func RegisterGatewayServer(addr string) {
	key := getGatewayServerPrefix() + addr
	registerWithLease(key, addr)
}

// RegisterGameServer 注册游戏服
func RegisterGameServer(serverID string, addr string) {
	key := getGameServerPrefix(serverID) + addr
	registerWithLease(key, addr)
}

// RegisterLogin 注册登录服
func RegisterLogin(addr string) {
	key := getLoginPrefix() + addr
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

func GetServiceList(serverName string) (list []string) {
	resp, err := etcdClient.Get(context.Background(), fmt.Sprintf("/%s/", serverName), clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	for _, kv := range resp.Kvs {
		list = append(list, string(kv.Value))
	}
	return
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

func UpdateEtcdConf(key, value string) {
	if !strings.HasPrefix(key, "/config/") {
		log.Printf("etcd更新配置参数异常: %s:%s", key, value)
		return
	}
	_, err := etcdClient.Put(context.Background(), key, value)
	if err != nil {
		log.Printf("etcd更新配置失败: %v", err)
		return
	}
}
