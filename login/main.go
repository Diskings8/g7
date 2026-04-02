package main

import (
	"flag"
	"g7/common/config"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/mysqlx"
	"g7/common/snowflake"
	"g7/common/utils"
	"g7/login/internal/dao_login"
	"g7/login/routers"
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

	logger.Init()
	logger.Log.Info("登录服启动中...")

	snowflake.Init()

	mysqlx.InitGlobalDb(config.GCfg.MySQLGlobal.Dsn())
	dao_login.AutoMigrate()

	r := routers.GetDefaultGin()
	routers.Register(r)

	logger.Log.Info("登录服启动绑定：" + config.GCfg.Server.Login)
	_ = r.Run(config.GCfg.Server.Login)
}
