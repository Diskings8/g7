package main

import (
	"flag"
	"g7/common/config"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/mysqlx"
	"g7/common/snowflake"
	"g7/common/utils"
	"g7/game/routers"
)

func main() {

	// 1. 解析环境参数
	flag.StringVar(&globals.Env, "env", "test", "运行环境: test/prod")
	flag.StringVar(&globals.ServerId, "server", "1001", "游戏服id")
	flag.Parse()

	// 2、获取配置
	var confStr string
	if !utils.IsDev() {
		confStr = globals.ConfPro
	} else {
		confStr = globals.ConfDev
	}
	config.Load(confStr)

	logger.Init()
	logger.Log.Info("游戏服启动中...")

	snowflake.Init()

	mysqlx.InitGlobalDb(config.GCfg.MySQLGlobal.Dsn())

	r := routers.GetDefaultGin()
	routers.Register(r)

	logger.Log.Info(globals.ServerId + " 游戏服启动绑定：" + config.GCfg.Server.Game)
	_ = r.Run(config.GCfg.Server.Game)
}
