package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
)

var (
	etcdClient *clientv3.Client
	once       sync.Once

	//GameServerCache 全局服务缓存（生产级）
	GameServerCache = &serviceCache{
		cache: make(map[string]string), // key: serverID, value: addr
	}

	GEtcdConfUpdateCenter = &etcdConfUpdateCenter{
		oneConfReloadCallbacks: make(map[string][]func()),
	}
)
