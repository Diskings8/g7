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
	"g7/common/redisx"
	"g7/common/snowflakes"
	"g7/login/global_login"
	"g7/login/internal/service_login"
	"g7/login/mq_login"
	"g7/login/routers"
	"os"
)

func main() {

	// 1. 解析环境参数
	flag.StringVar(&globals.Env, "env", "prod", "运行环境: test/prod")
	flag.StringVar(&globals.Platform, "platform", "91", "平台id")
	flag.StringVar(&globals.Container, "container", "docker", "容器类型：local/docker")
	flag.Parse()

	// 2、获取配置
	var confStr = globals.GetEnvConfPath()
	configx.LoadEnvConf(confStr)

	logger.Init()
	logger.Log.Info(fmt.Sprintf("本登录服启动：%s", configx.GEnvCfg.Server.Login))

	// 3、注册etcd,监听游戏服
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	etcd.GEtcdConfUpdateCenter.LoadAndWatchConfig()
	etcd.GEtcdConfUpdateCenter.RegisterConfReloadCallBack(etcd_conf.ConfSwitchLoginOn, func() {
		fmt.Println("Hello")
	})

	var etcdAddr string
	if globals.IsContainerDocker() {
		podIP := os.Getenv("POD_IP")
		//rpcPort := os.Getenv("RPC_PORT")
		Port := configx.GEnvCfg.Server.Login
		globals.InstanceId = os.Getenv("POD_NAME")
		etcdAddr = fmt.Sprintf("%s%s", podIP, Port)
		logger.Log.Info(fmt.Sprintf("本登录服%s 启动访问etcd", etcdAddr))
	} else {
		globals.InstanceId = "1"
		Port := configx.GEnvCfg.Server.Login
		etcdAddr = fmt.Sprintf("%s%s", "", Port)
		etcdAddr = fmt.Sprintf("%s", configx.GEnvCfg.Server.Login)
	}
	etcd.RegisterLoginRpc(globals.InstanceId, etcdAddr)

	//
	snowflakes.Init()

	//
	// 初始化redis
	redisx.Init(configx.GEnvCfg.Redis.Addr, configx.GEnvCfg.Redis.Password, configx.GEnvCfg.Redis.DB)

	//
	mq_login.GMQCustomInstance.Init()

	//
	service_login.LTServer.Init()

	// 4、使用数据库
	global_login.GLoginDB = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.Dsn())
	global_login.AutoMigrate(global_login.GLoginDB)

	// 5、初始化路由
	r := routers.GetDefaultGin()
	routers.Register(r)

	var serverAddr string
	serverAddr = configx.GEnvCfg.Server.Login
	logger.Log.Info(fmt.Sprintf("登录服启动绑定%s：%s", globals.Container, serverAddr))

	_ = r.Run(serverAddr)
}
