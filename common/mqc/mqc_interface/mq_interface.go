package mqc_interface

import "g7/common/model_common"

type MQProducerInterface interface {
	ProduceMessage(Topic string, Data model_common.DBMqInterface)
}

type MQConsumerInterface interface {
	RunConsumer()
}
