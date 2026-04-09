package kfk

import (
	"g7/common/model_common"
	"g7/common/mqc/mq_topic"
)

func topiToInstance(topic string) model_common.DBMqInterface {
	switch topic {
	case mq_topic.MakeGameActionTopicKey():
		return &model_common.ActionLog{}
	default:
		return &model_common.ActionLog{}
	}
}
