package main

import (
	"flag"
	"g7/common/config"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/mysqlx"
	"g7/common/protos/pb"
	"g7/common/snowflake"
	"g7/common/utils"
	"g7/game/rpc_server"
	"net"

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
	snowflake.Init()

	// 5、初始化数据库
	mysqlx.InitGlobalDb(config.GCfg.MySQLGlobal.Dsn())

	// 6、注册eetcd
	etcd.InitETCD(config.GCfg.Etcd.Dsn)
	etcd.RegisterGameServer(globals.ServerId, config.GCfg.Server.Game)

	// 7、注册grpc服务
	s := grpc.NewServer()
	pb.RegisterGameStreamServiceServer(s, &rpc_server.GameStreamServer{})
	pb.RegisterGameNodeServiceServer(s, &rpc_server.GameNodeServer{})

	logger.Log.Info(globals.ServerId + " 游戏服启动绑定：" + config.GCfg.Server.Game)
	lis, _ := net.Listen("tcp", config.GCfg.Server.Game)
	_ = s.Serve(lis)
}
