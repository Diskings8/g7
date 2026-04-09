package kfk

import (
	"encoding/json"
	"fmt"
	"g7/common/configx"
	"g7/common/dbc"
	"g7/common/dbc/dbc_interface"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/utils"
	"github.com/IBM/sarama"
	"log"
	"time"
)

type KafkaConsumerDriver struct {
	brokers         []string
	topic           string
	groupID         string // 消费者组
	consumerConn    sarama.Consumer
	batchMap        map[int32][]model_common.DBMqInterface // 按服分batch
	maxBatchSize    int
	commonDBConnMap map[int32]dbc_interface.DBInterface
	modelFunc       func() model_common.DBMqInterface
}

func NewKafkaConsumerDriver(brokers []string, topic, groupID string) *KafkaConsumerDriver {
	newModelFunc := func() model_common.DBMqInterface {
		return topiToInstance(topic)
	}
	dr := &KafkaConsumerDriver{
		brokers:         brokers,
		topic:           topic,
		groupID:         groupID,
		maxBatchSize:    200,
		modelFunc:       newModelFunc,
		batchMap:        make(map[int32][]model_common.DBMqInterface),
		commonDBConnMap: make(map[int32]dbc_interface.DBInterface),
	}
	if dr.Init() != nil {
		return nil
	}
	return dr
}

// Init 初始化消费者
func (cd *KafkaConsumerDriver) Init() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true // 开启错误返回

	// 连接消费者
	consumer, err := sarama.NewConsumer(cd.brokers, config)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}
	cd.consumerConn = consumer
	return nil
}

func (cd *KafkaConsumerDriver) RunConsumer() {
	// 1. 获取该 Topic 的所有分区（修复点：不再写死 0）
	partitions, err := cd.consumerConn.Partitions(cd.topic)
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	// 2. 遍历所有分区启动消费
	for _, partition := range partitions {
		go cd.consumePartition(partition)
	}

	// 保持进程不退出
	select {}
}

// 消费单个分区
func (cd *KafkaConsumerDriver) consumePartition(partition int32) {
	log.Printf("%s Start consuming partition: %d", cd.topic, partition)

	// 从最新的 offset 开始
	offset := sarama.OffsetNewest
	consumer, err := cd.consumerConn.ConsumePartition(cd.topic, partition, offset)
	if err != nil {
		log.Fatalf("Failed to consume partition %d: %v", partition, err)
	}
	defer consumer.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-consumer.Messages():
			// 解析消息
			mqLog := cd.modelFunc()
			if err := json.Unmarshal(msg.Value, mqLog); err != nil {
				logger.Log.Warn(fmt.Sprintf("Unmarshal error: %v, msg: %s", err, string(msg.Value)))
				continue // 解析失败的消息跳过，或者入死信队列
			}

			// 区分区服内容
			serverId := mqLog.GetServerId()
			// 加入批量
			cd.batchMap[serverId] = append(cd.batchMap[serverId], mqLog)

			// 批量插入逻辑
			if len(cd.batchMap[serverId]) >= cd.maxBatchSize {
				if errSave := cd.saveBatch(serverId, cd.batchMap[serverId]); errSave != nil {
					log.Printf("Batch insert failed: %v", errSave)
					// 失败不清空batch，下次继续尝试
					continue
				}
				// 成功后才清空
				cd.batchMap[serverId] = cd.batchMap[serverId][:0]

			}

		case <-ticker.C:
			// 定时批量插入
			for sid, batch := range cd.batchMap {
				if len(batch) > 0 {
					if errSave := cd.saveBatch(sid, batch); errSave != nil {
						log.Printf("Batch insert failed: %v", errSave)
						// 失败不清空batch，下次继续尝试
						continue
					}
					cd.batchMap[sid] = cd.batchMap[sid][:0]
				}
			}

		case err := <-consumer.Errors():
			log.Printf("Kafka consumer error: %v", err)
		}
	}
}

func (m *KafkaConsumerDriver) getConn(serverId int32) dbc_interface.DBInterface {
	v, ok := m.commonDBConnMap[serverId]
	if ok {
		return v
	}
	c := dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.DsnWithName(utils.Int32ToString(serverId)))
	m.commonDBConnMap[serverId] = c
	return c
}

// saveBatch 批量插入数据库
func (cd *KafkaConsumerDriver) saveBatch(serverId int32, batch []model_common.DBMqInterface) error {
	if len(batch) == 0 {
		return nil
	}
	defer func() {
		if err := recover(); err != nil {
			logger.Log.Error(fmt.Sprintf("saveBatch recover err: %v", err))
		}
	}()
	db := cd.getConn(serverId)
	if !db.IsTableExists(batch[0].TableName()) {
		if err := db.AutoMigrate(batch[0]); err != nil {
			return err
		}
	}

	// 这里加个事务保护，确保批量插入成功
	tx := db.Begin()
	if tx == nil {
		return fmt.Errorf("failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error(fmt.Sprintf("tx panic: %v", r))
			_ = tx.Rollback()
		}
	}()

	if err := tx.BatchMQInsert(batch); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
