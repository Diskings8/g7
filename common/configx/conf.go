package configx

import (
	"g7/common/configx/env_conf"
	"g7/common/configx/etcd_conf"
	"gopkg.in/yaml.v3"
	"os"
)

var GEnvCfg env_conf.Config
var GEtcdCfg etcd_conf.Config

// LoadEnvConf 加载启动配置文件
func LoadEnvConf(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("load configx fail: " + err.Error())
	}

	err = yaml.Unmarshal(data, &GEnvCfg)
	if err != nil {
		panic("parse configx fail: " + err.Error())
	}
}
