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
	"g7/comprehensive/rpc_server/match_server"
	"g7/comprehensive/rpc_server/room_server"
	"g7/comprehensive/rpc_server/roommanager_server"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
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
	var etcdMatchAddr, etcdRoomManagerAddr, etcdRoomAddr string
	if globals.IsContainerDocker() {
		globals.InstanceId = os.Getenv("POD_NAME")
		podIP := os.Getenv("POD_IP")
		etcdMatchAddr = fmt.Sprintf("%s%s", podIP, configx.GEnvCfg.Comprehensive.Match)
		etcdRoomManagerAddr = fmt.Sprintf("%s%s", podIP, configx.GEnvCfg.Comprehensive.RoomManager)
		etcdRoomAddr = fmt.Sprintf("%s%s", podIP, configx.GEnvCfg.Comprehensive.Room)
	} else {
		globals.InstanceId = "1"
		etcdMatchAddr = fmt.Sprintf("%s%s", "", configx.GEnvCfg.Comprehensive.Match)
		etcdRoomManagerAddr = fmt.Sprintf("%s%s", "", configx.GEnvCfg.Comprehensive.RoomManager)
		etcdRoomAddr = fmt.Sprintf("%s%s", "", configx.GEnvCfg.Comprehensive.Room)
	}
	etcd.RegisterMatchNodeRpc(globals.InstanceId, etcdMatchAddr)
	etcd.RegisterRoomManagerNodeRpc(globals.InstanceId, etcdRoomManagerAddr)
	etcd.RegisterRoomNodeRpc(globals.InstanceId, etcdRoomAddr)

	//初始化定时器
	cronx.InitCron()

	//初始化管理系
	manager_system.GMatchManager.Init()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 注册grpc服务
	var (
		matchServer   *grpc.Server
		roomMgrServer *grpc.Server
		roomServer    *grpc.Server
	)

	matchServer = runMatchServer()
	roomMgrServer = runRoomManagerServer()
	roomServer = runRoomServer()

	<-sigChan
	logger.Log.Info("收到退出信号，开始关闭服务...")
	if roomServer != nil {
		roomServer.GracefulStop()

	}
	if roomMgrServer != nil {
		roomMgrServer.GracefulStop()

	}
	if matchServer != nil {
		matchServer.GracefulStop()
		
	}
	logger.Log.Info("所有服务已退出，进程结束")
}

func runMatchServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterMatchNodeServiceServer(s, &match_server.MatchServer{})
	var serverAddr string
	serverAddr = configx.GEnvCfg.Comprehensive.Match
	logger.Log.Info(fmt.Sprintf("公共服%s启动绑定%s：%s", "runMatchServer", globals.Container, serverAddr))

	lis, _ := net.Listen("tcp", serverAddr)
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Log.Error(fmt.Sprintf("runMatchServer server error: %v", err))
		}
	}()
	return s
}

func runRoomManagerServer() *grpc.Server {
	s := grpc.NewServer()

	var serverAddr string
	serverAddr = configx.GEnvCfg.Comprehensive.RoomManager
	logger.Log.Info(fmt.Sprintf("公共服%s启动绑定%s：%s", "runRoomManagerServer", globals.Container, serverAddr))

	roommanager_server.GRoomManagerServer.Init()
	pb.RegisterRoomManagerNodeServiceServer(s, roommanager_server.GRoomManagerServer)

	lis, _ := net.Listen("tcp", serverAddr)
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Log.Error(fmt.Sprintf("runRoomManagerServer server error: %v", err))
		}
	}()
	return s
}

func runRoomServer() *grpc.Server {
	s := grpc.NewServer()

	var serverAddr string
	serverAddr = configx.GEnvCfg.Comprehensive.Room
	logger.Log.Info(fmt.Sprintf("公共服%s启动绑定%s：%s", "runRoomServer", globals.Container, serverAddr))

	room_server.GRoomMixServer.Init(serverAddr)
	pb.RegisterRoomNodeServiceServer(s, room_server.GRoomMixServer)
	pb.RegisterRoomStreamServiceServer(s, room_server.GRoomMixServer)

	lis, _ := net.Listen("tcp", serverAddr)
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Log.Error(fmt.Sprintf("Room server error: %v", err))
		}
	}()
	return s
}
