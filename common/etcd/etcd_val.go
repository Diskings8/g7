package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
)

var (
	etcdClient *clientv3.Client
	once       sync.Once
	
	GEtcdConfUpdateCenter = &etcdConfUpdateCenter{
		oneConfReloadCallbacks: make(map[string][]func()),
	}
)
