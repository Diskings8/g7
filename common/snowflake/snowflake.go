package snowflake

import (
	"g7/common/config"
	"g7/common/logger"
	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

var node *snowflake.Node

func Init() {
	// 节点ID 1~1023，专服可以配置
	// 从 YAML 读取配置
	cfg := config.GCfg.Snowflake

	// 设置雪花算法参数
	snowflake.NodeBits = 10
	snowflake.StepBits = 12

	// 创建节点：使用 datacenter + worker 组合
	nodeID := (cfg.DatacenterID << 5) | cfg.WorkerID

	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		logger.Log.Fatal("雪花算法初始化失败", zap.Error(err))
		panic(err)
	}
	node = n

	logger.Log.Info("雪花算法初始化成功",
		zap.Int64("datacenter_id", cfg.DatacenterID),
		zap.Int64("worker_id", cfg.WorkerID),
		zap.Int64("node_id", nodeID),
	)
}

// GenUID 生成全局唯一UID
func GenUID() int64 {
	return node.Generate().Int64()
}
