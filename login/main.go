package main

import (
	"flag"
	"fmt"
	"g7/common/configx"
	"g7/common/configx/etcd_conf"
	"g7/common/dbc"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/snowflakes"
	"g7/common/utils"
	"g7/login/global_login"
	"g7/login/model_login"
	"g7/login/mq_login"
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
	configx.LoadEnvConf(confStr)

	logger.Init()
	logger.Log.Info("登录服启动中...")

	// 3、注册etcd,监听游戏服
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	etcd.GEtcdConfUpdateCenter.LoadAndWatchConfig()
	etcd.GEtcdConfUpdateCenter.RegisterConfReloadCallBack(etcd_conf.ConfSwitchLoginOn, func() {
		fmt.Println("Hello")
	})

	etcd.RegisterLogin(configx.GEnvCfg.Server.Login)

	//
	snowflakes.Init()

	//
	mq_login.GMQCustomInstance.Init()

	// 4、使用数据库
	global_login.GLoginDB = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.Dsn())
	_ = dbc.AutoMigrates(global_login.GLoginDB, &model_login.User{})

	// 5、初始化路由
	r := routers.GetDefaultGin()
	routers.Register(r)

	logger.Log.Info("登录服启动绑定：" + configx.GEnvCfg.Server.Login)
	_ = r.Run(configx.GEnvCfg.Server.Login)
}
