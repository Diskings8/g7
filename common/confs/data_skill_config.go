package confs

import (
	"encoding/json"
	"os"
	"sync"
)

/*
    脚本生成,请勿修改
*/

// SkillConfig 配置结构体
type SkillConfig struct {
    Id int32 `json:"id"` // 技能id
    Name string `json:"name"` // 技能名称
    Skillcost float64 `json:"SkillCost"` // 能量消耗
    Skilltype int32 `json:"SkillType"` // 技能类型
    Targettype int32 `json:"TargetType"` // 目标对象类型
    Score float64 `json:"Score"` // 数值
    Cd float64 `json:"CD"` // 冷却
    Range float64 `json:"Range"` // 范围
}

// ConfigSkill 配置管理结构体
type ConfigSkill struct {
	RWLock  sync.RWMutex
	DataMap map[int32]*SkillConfig
}

// LoadConfig 加载配置到内存
func (c *ConfigSkill) LoadConfig(path string) error {
	data, err := os.ReadFile(path+"/data_skill_config.json")
	if err != nil {
		return err
	}

	var list []SkillConfig
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	c.RWLock.Lock()
	defer c.RWLock.Unlock()

	c.DataMap = make(map[int32]*SkillConfig)
	for _, v := range list {
		c.DataMap[v.Id] = &v
	}

	return nil
}

// Find 根据ID获取配置（并发安全）
func (c *ConfigSkill) Find(id int32) (*SkillConfig, bool) {
	c.RWLock.RLock()
	defer c.RWLock.RUnlock()

	v, ok := c.DataMap[id]
	return v, ok
}
