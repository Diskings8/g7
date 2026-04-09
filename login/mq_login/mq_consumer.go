package mq_login

import (
	"g7/common/configx"
	"g7/common/mqc"
	"g7/common/mqc/mq_topic"
)

var GMQCustomInstance MQCustomInstance

type MQCustomInstance struct {
	topicNumMap map[string]int // topic对应的消费者数量
}

func (s *MQCustomInstance) Init() {
	s.topicNumMap = make(map[string]int)
	s.topicNumMap[mq_topic.MakeGameCreateRoleTopicKey()] = 1
	s.topicNumMap[mq_topic.MakeGameActionTopicKey()] = 1

	s.initAllCustomer()
}

func (s *MQCustomInstance) initAllCustomer() {
	url := configx.GEnvCfg.MQ.Dsn
	kind := configx.GEnvCfg.MQ.Kind
	for topic, num := range s.topicNumMap {
		for i := 0; i < num; i++ {
			topicConsumer := mqc.InitMQConsumer(kind, topic, url)
			go topicConsumer.RunConsumer()
		}
	}
}
