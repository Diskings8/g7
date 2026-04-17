package etcd

import (
	"context"
	"fmt"
	"g7/common/globals"
	"g7/common/structs"
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

func getGatewayTcpPrefix() string {
	return "/" + globals.GatewayTcp + "/"
}

func getGatewayRpcPrefix() string {
	return "/" + globals.GatewayRpc + "/"
}

func GetAllGameRpcPrefix() string {
	return "/" + globals.GameRpc + "/"
}

func getOneKindGameRpcPrefix(serverID string) string {
	return "/" + globals.GameRpc + "/" + serverID
}

func getLoginRpcPrefix() string {
	return "/" + globals.LoginRpc + "/"
}

// RegisterGatewayTcp 注册网关
func RegisterGatewayTcp(instance, addr string) {
	key := getGatewayTcpPrefix() + instance + "/" + addr
	registerWithLease(key, addr)
}

func RegisterGatewayRpc(instance, addr string) {
	key := getGatewayRpcPrefix() + instance + "/" + addr
	registerWithLease(key, addr)
}

/*
RegisterGameRpc 注册游戏服
eg: /game_rpc/91001_0/172.0.0.4:8080
*/
func RegisterGameRpc(serverID, instance string, addr string) {
	key := getOneKindGameRpcPrefix(serverID) + "_" + instance + "/" + addr
	registerWithLease(key, addr)
}

// RegisterLoginRpc 注册登录服
func RegisterLoginRpc(instance, addr string) {
	key := getLoginRpcPrefix() + instance + "/" + addr
	registerWithLease(key, addr)
}

// 内部通用：带租约注册
func registerWithLease(key, value string) {
	fmt.Println(key, value)
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

// ShowServiceList 制作展示用途
func ShowServiceList(serverName string) (list []string) {
	resp, err := etcdClient.Get(context.Background(), fmt.Sprintf("/%s/", serverName), clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	for _, kv := range resp.Kvs {
		list = append(list, fmt.Sprintf("{%s # %s}", string(kv.Key), string(kv.Value)))
	}
	return
}

// GetAllGateways 获取所有网关
func GetAllGateways() ([]string, error) {
	resp, err := etcdClient.Get(context.Background(), getGatewayTcpPrefix(), clientv3.WithPrefix())
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
func GetGameServersByServerID(serverID string) ([]structs.KVString, error) {
	key := getOneKindGameRpcPrefix(serverID)
	resp, err := etcdClient.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var addrs []structs.KVString
	for _, kv := range resp.Kvs {
		addrs = append(addrs, structs.KVString{string(kv.Key), string(kv.Value)})
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
