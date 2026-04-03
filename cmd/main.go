package main

import (
	"flag"
	"g7/common/config"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/pkg/mysqlx"
	"g7/common/redisx"
	"g7/common/snowflake"
	"g7/common/utils"
	"go.uber.org/zap"
	"strconv"
)

func main() {
	// 1. 解析环境参数
	flag.StringVar(&globals.Env, "env", "test", "运行环境: test/prod")
	flag.Parse()

	// 2、获取配置
	var confStr string
	if !utils.IsDev() {
		confStr = globals.ConfPro
	} else {
		confStr = globals.ConfDev
	}
	config.Load(confStr)

	// 3. 初始化日志
	logger.Init()
	logger.Log.Info("启动服务", zap.String("env", globals.Env))

	// 4. 初始化 MySQL
	mysqlx.Init(config.Cfg.MySQL.DSN)

	// 5. 初始化 Redis
	redisx.Init(config.Cfg.Redis.Addr, config.Cfg.Redis.Password, config.Cfg.Redis.DB)

	// 6. 初始化雪花算法（从YAML读取）
	snowflake.Init()

	// 8. 启动
	port := config.Cfg.Server.Port
	logger.Log.Info("服务启动成功", zap.Int("port", port))
	_ = r.Run(":" + strconv.Itoa(port))
}
