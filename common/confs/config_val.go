package confs

import (
    "fmt"
    "g7/common/configx"
)

/*
    脚本生成,请勿修改
*/

var (
    GConfigDataItem = ConfigDataItem{}
)

// ReloadAllConfig 统一加载/热重载所有配置
func ReloadAllConfig() error {
    var path = configx.GEnvCfg.JsonPath.Path

    if err := GConfigDataItem.LoadConfig(path); err != nil {
        fmt.Printf("GConfigDataItem.LoadError: %s",err.Error())
    }
    return nil
}
