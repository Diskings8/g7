package kfk

import (
	"g7/common/model_common"
	"g7/common/mqc/mq_topic"
	"time"
)

func topiToInstance(topic string) model_common.DBMqInterface {
	switch topic {
	case mq_topic.MakeGameActionTopicKey():
		d := &model_common.ActionLog{}
		d.SaveTime = time.Now().Unix()
		return d
	case mq_topic.MakeGameCreateRoleTopicKey():
		d := &model_common.ActionLog{}  // 角色创建也使用ActionLog模型
		d.SaveTime = time.Now().Unix()
		return d
	default:
		d := &model_common.ActionLog{}
		d.SaveTime = time.Now().Unix()
		return d
	}
}
