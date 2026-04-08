package main

import (
	"flag"
	"g7/common/config"
	"g7/common/cronx"
	"g7/common/dbc"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/snowflakes"
	"g7/common/utils"
	"g7/game/global_game"
	"g7/game/manager_game"
	"g7/game/rpc_server"
	"net"

	_ "g7/game/activity_system_game"
	_ "g7/game/cultivation_system_game"
	_ "g7/game/general_system_game"

	"google.golang.org/grpc"
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

	// 3、初始化日志
	logger.Init()
	logger.Log.Info("游戏服启动中...")

	// 4、初始化雪花
	snowflakes.Init()

	// 5、初始化数据库
	global_game.GGameDB = dbc.InitDB(globals.DBMysql, config.GCfg.MySQLGame.DsnWithName(globals.ServerId))
	global_game.GGlobalDB = dbc.InitDB(globals.DBMysql, config.GCfg.MySQLGlobal.Dsn())
	global_game.AutoMigrate(global_game.GGameDB)

	// 初始化redis
	redisx.Init(config.GCfg.Redis.Addr, config.GCfg.Redis.Password, config.GCfg.Redis.DB)

	// 注册etcd
	etcd.InitETCD(config.GCfg.Etcd.Dsn)
	etcd.RegisterGameServer(globals.ServerId, config.GCfg.Server.Game)

	//全局对象初始化 init
	global_game.GPlayerMaps.Init()
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

	logger.Log.Info(globals.ServerId + " 游戏服启动绑定：" + config.GCfg.Server.Game)
	lis, _ := net.Listen("tcp", config.GCfg.Server.Game)
	_ = s.Serve(lis)
}
