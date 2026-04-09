package etcd

import (
	"context"
	"g7/common/configx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdConfUpdateCenter struct {
	currentVersion         int64
	oneConfReloadCallbacks map[string][]func()
	confReloadCallbacks    []func()
}

func (rb *etcdConfUpdateCenter) RegisterConfReloadCallBack(confKey string, rf func()) {
	rb.oneConfReloadCallbacks[confKey] = append(rb.oneConfReloadCallbacks[confKey], rf)
}
func (rb *etcdConfUpdateCenter) RegisterAllConfReloadCallBack(rf func()) {
	rb.confReloadCallbacks = append(rb.confReloadCallbacks, rf)
}

func (rb *etcdConfUpdateCenter) LoadAndWatchConfig() {
	// 1. 拉取全量配置
	rb.loadAllConfig()
	// 2. 监听配置变化（热更新）
	rb.watchConfig()
}

func (rb *etcdConfUpdateCenter) loadAllConfig() {
	resp, _ := etcdClient.Get(context.Background(), "/config/", clientv3.WithPrefix())

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)
		configx.GEtcdCfg.SetConf(key, value)
	}
}

func (rb *etcdConfUpdateCenter) watchConfig() {
	watchChan := etcdClient.Watch(context.Background(), "/config/", clientv3.WithPrefix())

	go func() {
		for resp := range watchChan {
			var updateKeys []string
			for _, ev := range resp.Events {
				key := string(ev.Kv.Key)
				updateKeys = append(updateKeys, key)
				value := string(ev.Kv.Value)
				configx.GEtcdCfg.SetConf(key, value)
			}
			for _, key := range updateKeys {
				for _, fc := range rb.oneConfReloadCallbacks[key] {
					fc()
				}
			}
			for _, f := range rb.confReloadCallbacks {
				f() //
			}
		}

	}()
}
