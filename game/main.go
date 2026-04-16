package main

import (
	"flag"
	"fmt"
	"g7/common/configx"
	"g7/common/confs"
	"g7/common/cronx"
	"g7/common/dbc"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/mqc"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/snowflakes"
	"g7/game/global_game"
	"g7/game/manager_game"
	"g7/game/rpc_server"
	"net"
	"os"

	_ "g7/game/activity_system_game"
	_ "g7/game/cultivation_system_game"
	_ "g7/game/general_system_game"

	"google.golang.org/grpc"
)

func main() {
	// 1. 解析环境参数
	flag.StringVar(&globals.Env, "env", "prod", "运行环境: test/prod")
	flag.StringVar(&globals.ServerId, "server", "1001", "游戏服id")
	flag.StringVar(&globals.Platform, "platform", "91", "平台id")
	flag.StringVar(&globals.Container, "container", "docker", "容器类型：local/docker")

	flag.Parse()

	// 2、获取配置
	var confStr = globals.GetEnvConfPath()
	configx.LoadEnvConf(confStr)

	// 3、初始化日志
	logger.Init()
	logger.Log.Info(fmt.Sprintf("游戏服%s 启动中...", globals.ServerId))

	//
	_ = confs.ReloadAllConfig()

	// 4、初始化雪花
	snowflakes.Init()
	//logger.Log.Info(fmt.Sprintf("数据库%s", configx.GEnvCfg.MySQLGame.DsnWithName(globals.ServerId)))
	// 5、初始化数据库
	global_game.GGameDB = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGame.DsnWithName(globals.ServerId))
	global_game.GGlobalDB = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.Dsn())
	global_game.AutoMigrate(global_game.GGameDB)

	// 初始化redis
	redisx.Init(configx.GEnvCfg.Redis.Addr, configx.GEnvCfg.Redis.Password, configx.GEnvCfg.Redis.DB)

	// 初始化mq
	global_game.GGlobalMQ = mqc.InitMQProducer(configx.GEnvCfg.MQ.Kind, configx.GEnvCfg.MQ.Dsn)

	// 注册etcd
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	etcd.GEtcdConfUpdateCenter.LoadAndWatchConfig()
	var etcdAddr string
	if globals.IsContainerDocker() {
		podIP := os.Getenv("POD_IP")
		Port := configx.GEnvCfg.Server.Game
		globals.InstanceId = os.Getenv("POD_NAME")
		etcdAddr = fmt.Sprintf("%s%s", podIP, Port)
		logger.Log.Info(fmt.Sprintf("本登录服%s 启动访问etcd：%s", globals.InstanceId, etcdAddr))
	} else {
		globals.InstanceId = "1"
		Port := configx.GEnvCfg.Server.Game
		etcdAddr = fmt.Sprintf("%s%s", "", Port)
		etcdAddr = fmt.Sprintf("%s", configx.GEnvCfg.Server.Game)
	}
	etcd.RegisterGameRpc(globals.ServerId, globals.InstanceId, etcdAddr)

	//全局对象初始化 init
	global_game.GPlayerMaps.Init(redisx.GetClient())
	global_game.GPlayerCache.Init()

	//初始化定时器
	cronx.InitCron()

	//初始化管理系
	manager_game.GSaveSystemManager.Init()
	manager_game.GPlayerManager.Init()

	// 注册grpc服务
	s := grpc.NewServer()
	pb.RegisterGameStreamServiceServer(s, &rpc_server.GameStreamServer{})
	pb.RegisterGameNodeServiceServer(s, &rpc_server.GameNodeServer{})

	var serverAddr string
	serverAddr = configx.GEnvCfg.Server.Game
	logger.Log.Info(fmt.Sprintf("游戏%s服启动绑定%s：%s", globals.ServerId, globals.Container, serverAddr))

	lis, _ := net.Listen("tcp", serverAddr)
	_ = s.Serve(lis)
}
