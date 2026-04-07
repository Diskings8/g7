package main

import (
	"flag"
	"g7/common/config"
	"g7/common/dbc"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/snowflakes"
	"g7/common/utils"
	"g7/login/global_login"
	"g7/login/model_login"
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

	// 3、注册etcd,监听游戏服
	etcd.InitETCD(config.GCfg.Etcd.Dsn)
	etcd.RegisterLogin(config.GCfg.Server.Login)

	snowflakes.Init()

	// 4、使用数据库
	global_login.GLoginDB = dbc.InitDB(globals.DBMysql, config.GCfg.MySQLGlobal.Dsn())
	_ = dbc.AutoMigrates(global_login.GLoginDB, &model_login.User{})

	// 5、初始化路由
	r := routers.GetDefaultGin()
	routers.Register(r)

	logger.Log.Info("登录服启动绑定：" + config.GCfg.Server.Login)
	_ = r.Run(config.GCfg.Server.Login)
}
