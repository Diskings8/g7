package main

import (
	"flag"
	"fmt"
	"g7/common/configx"
	"g7/common/cronx"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/comprehensive/manager_system"
	"g7/comprehensive/rpc_server"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main() {
	// 1. 解析环境参数
	flag.StringVar(&globals.Env, "env", "test", "运行环境: test/prod")
	flag.StringVar(&globals.Platform, "platform", "91", "平台id")
	flag.StringVar(&globals.Container, "container", "docker", "容器类型：local/docker")

	flag.Parse()

	// 2、获取配置
	var confStr = globals.GetEnvConfPath()
	configx.LoadEnvConf(confStr)

	// 3、初始化日志
	logger.Init()
	logger.Log.Info(fmt.Sprintf("综合服%s 启动中...", globals.ServerId))

	// 初始化redis
	redisx.Init(configx.GEnvCfg.Redis.Addr, configx.GEnvCfg.Redis.Password, configx.GEnvCfg.Redis.DB)

	// 初始化mq
	//global_game.GGlobalMQ = mqc.InitMQProducer(configx.GEnvCfg.MQ.Kind, configx.GEnvCfg.MQ.Dsn)

	// 注册etcd
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	etcd.GEtcdConfUpdateCenter.LoadAndWatchConfig()
	var etcdMatchAddr, etcdRoomManagerAddr string
	if globals.IsContainerDocker() {
		globals.InstanceId = os.Getenv("POD_NAME")
		podIP := os.Getenv("POD_IP")
		etcdMatchAddr = fmt.Sprintf("%s%s", podIP, configx.GEnvCfg.Comprehensive.Match)
		etcdRoomManagerAddr = fmt.Sprintf("%s%s", podIP, configx.GEnvCfg.Comprehensive.RoomManager)
	} else {
		globals.InstanceId = "1"
		etcdMatchAddr = fmt.Sprintf("%s%s", "", configx.GEnvCfg.Comprehensive.Match)
		etcdRoomManagerAddr = fmt.Sprintf("%s%s", "", configx.GEnvCfg.Comprehensive.RoomManager)
	}
	etcd.RegisterMatchNodeRpc(globals.InstanceId, etcdMatchAddr)
	etcd.RegisterRoomManagerNodeRpc(globals.InstanceId, etcdRoomManagerAddr)

	//初始化定时器
	cronx.InitCron()

	//初始化管理系
	manager_system.GMatchManager.Init()

	// 注册grpc服务
	s := grpc.NewServer()
	pb.RegisterMatchNodeServiceServer(s, &rpc_server.MatchServer{})

	var serverAddr string
	serverAddr = configx.GEnvCfg.Comprehensive.Match
	logger.Log.Info(fmt.Sprintf("公共服启动绑定%s：%s", globals.Container, serverAddr))

	lis, _ := net.Listen("tcp", serverAddr)
	_ = s.Serve(lis)
}
