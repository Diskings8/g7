package main

import (
	"fmt"
	"g7/common/config"
	"g7/common/globals"
	"g7/common/model_common"
	"g7/common/mqc"
	"g7/common/structs"
	"os"
	"time"
)

func main() {
	fmt.Println(os.Getwd())
	config.Load(globals.ConfDev)
	url := "127.0.0.1:9092"
	producer := mqc.InitMQProducer(config.GCfg.MQ.Kind, config.GCfg.MQ.Dsn)

	topic := mqc.MakeGameActionTopicKey()
	consum := mqc.InitMQConsumer("kafka", topic, url, nil)
	go consum.RunConsumer()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	// 无限循环监听
	for {
		<-ticker.C
		data := &model_common.ActionLog{
			PlayerID:     2041413406195585024,
			Action:       "tst",
			Reason:       "test kafka",
			CostItem:     []structs.KInt32VInt64{{K: 1, V: 1}},
			CostCurrency: []structs.KInt32VInt64{{K: 1001, V: 1}},
			GainItem:     []structs.KInt32VInt64{{K: 2, V: 1}},
			GainCurrency: []structs.KInt32VInt64{{K: 1002, V: 10}},
			Ext:          "",
		}
		data.CreateTime = time.Now()
		producer.ProduceMessage(topic, data)
	}
}
