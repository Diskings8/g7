package confs

import (
	"encoding/json"
	"os"
	"sync"
)

/*
    脚本生成,请勿修改
*/

// DataItemConfig 配置结构体
type DataItemConfig struct {
    Id int32 `json:"id"` // 道具id
    Name string `json:"name"` // 道具名称
    Resourcetype int32 `json:"ResourceType"` // 道具类型
    Resourcesubtype int32 `json:"ResourceSubType"` // 道具子类型
    Price int32 `json:"price"` // 道具价格
    Isbind int32 `json:"IsBind"` // 是否绑定
    Isunique int32 `json:"IsUnique"` // 是否唯一
    Expiretype int32 `json:"ExpireType"` // 过期类型
    Limittime int64 `json:"LimitTime"` // 过期时间
}

// ConfigDataItem 配置管理结构体
type ConfigDataItem struct {
	RWLock  sync.RWMutex
	DataMap map[int32]*DataItemConfig
}

// LoadConfig 加载配置到内存
func (c *ConfigDataItem) LoadConfig(path string) error {
	data, err := os.ReadFile(path+"/data_item_config.json")
	if err != nil {
		return err
	}

	var list []DataItemConfig
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	c.RWLock.Lock()
	defer c.RWLock.Unlock()

	c.DataMap = make(map[int32]*DataItemConfig)
	for _, v := range list {
		c.DataMap[v.Id] = &v
	}

	return nil
}

// Find 根据ID获取配置（并发安全）
func (c *ConfigDataItem) Find(id int32) (*DataItemConfig, bool) {
	c.RWLock.RLock()
	defer c.RWLock.RUnlock()

	v, ok := c.DataMap[id]
	return v, ok
}
