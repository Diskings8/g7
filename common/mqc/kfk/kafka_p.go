package kfk

import (
	"encoding/json"
	"g7/common/model_common"
	"github.com/IBM/sarama"
	"log"
	"time"
)

var temp = "127.0.0.1:9092"

type KafkaProducerDriver struct {
	producerClient sarama.AsyncProducer
}

func NewKafkaProducerDriver(url string) *KafkaProducerDriver {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // 生产级别可靠性
	config.Producer.Compression = sarama.CompressionSnappy   // 压缩
	config.Producer.Flush.Frequency = 500 * time.Millisecond // 批量发送
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer([]string{url}, config)
	if err != nil {
		log.Fatalf("kafka启动失败: %v", err)
		return nil
	} else {
		go func() {
			for err := range producer.Errors() {
				log.Printf("kafka发送失败: %v", err)
			}
		}()
	}
	//log.Println("Kafka 异步生产者启动成功")
	return &KafkaProducerDriver{producerClient: producer}
}

func (k *KafkaProducerDriver) ProduceMessage(Topic string, Data model_common.DBMqInterface) {
	bytes, _ := json.Marshal(Data)
	log.Printf("one topic:%s", Topic)
	// 发送到kafka（异步，无阻塞）
	k.producerClient.Input() <- &sarama.ProducerMessage{
		Topic: Topic, // 主题固定
		Value: sarama.ByteEncoder(bytes),
	}
}
