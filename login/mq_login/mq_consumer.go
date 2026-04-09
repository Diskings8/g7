package mq_login

import (
	"g7/common/config"
	"g7/common/mqc"
)

type MQCustomInstance struct {
	topicNumMap map[string]int // topic对应的消费者数量
}

func (s *MQCustomInstance) Init() {
	s.topicNumMap = make(map[string]int)
	s.topicNumMap[mqc.MakeGameCreateRoleTopicKey()] = 1
	s.topicNumMap[mqc.MakeGameActionTopicKey()] = 2

	s.initAllCustomer()
}

func (s *MQCustomInstance) initAllCustomer() {
	url := config.GCfg.MQ.Dsn
	kind := config.GCfg.MQ.Kind
	for topic, num := range s.topicNumMap {
		for i := 0; i < num; i++ {
			topicConsumer := mqc.InitMQConsumer(kind, topic, url)
			go topicConsumer.RunConsumer()
		}
	}
}
