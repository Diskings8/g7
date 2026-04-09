package mqc

import (
	"g7/common/dbc/dbc_interface"
	"g7/common/mqc/kfk"
	"g7/common/mqc/mqc_interface"
)

func InitMQProducer(mqType, url string) mqc_interface.MQProducerInterface {
	switch mqType {
	case "kafka":
		kfkPD := kfk.NewKafkaProducerDriver(url)
		return kfkPD
	default:
		return nil
	}
}

func InitMQConsumer(mqType, topic, url string) mqc_interface.MQConsumerInterface {
	switch mqType {
	case "kafka":
		kfkCD := kfk.NewKafkaConsumerDriver([]string{url}, topic, "common")
		return kfkCD
	default:
		return nil
	}
}
