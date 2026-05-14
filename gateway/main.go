package main

import (
	"context"
	"flag"
	"fmt"
	"g7/common/configx"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/gateway/global_gateway"
	"g7/gateway/mix_server"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	// 解析环境参数
	flag.StringVar(&globals.Env, "env", "prod", "运行环境: test/prod/pre")
	//flag.StringVar(&globals.InstanceId, "instance", "1", "实例id")
	flag.StringVar(&globals.Platform, "platform", "91", "平台id")
	flag.StringVar(&globals.Container, "container", "docker", "容器类型：local/docker")
	flag.Parse()

	globals.InitSysSigChan()

	// 获取环境配置
	var confStr = globals.GetEnvConfPath()
	configx.LoadEnvConf(confStr)
	log.Println("网关加载配置完成")

	//
	logger.Init()
	logger.Log.Info(fmt.Sprintf("本网关启动配置：%s", confStr))

	go func() {
		// 端口随便选，不和你的服务冲突即可 比如 6061
		err := http.ListenAndServe("0.0.0.0:6061", nil)
		if err != nil {
			panic(err)
		}
	}()

	// 注册etcd,监听游戏服
	var etcdTcpAddr, etcdRpcAddr string
	if globals.IsContainerDocker() {
		globals.InstanceId = os.Getenv("POD_NAME")
		podIP := os.Getenv("POD_IP")
		//rpcPort := os.Getenv("RPC_PORT")
		tcpPort := configx.GEnvCfg.GateWay.Port
		rpcPort := configx.GEnvCfg.GateWay.RpcPort
		etcdTcpAddr = fmt.Sprintf("%s:%s", podIP, tcpPort)
		etcdRpcAddr = fmt.Sprintf("%s:%s", podIP, rpcPort)
	} else {
		globals.InstanceId = "1"
		etcdTcpAddr = fmt.Sprintf("%s", configx.GEnvCfg.GateWay.Dsn())
		etcdRpcAddr = fmt.Sprintf("%s", configx.GEnvCfg.GateWay.RpcDsn())
	}

	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)

	etcd.GEtcdConfUpdateCenter.LoadAndWatchConfig()
	logger.Log.Info("网关监听etcd完成")

	etcd.RegisterGatewayTcp(globals.InstanceId, etcdTcpAddr)
	etcd.RegisterGatewayRpc(globals.InstanceId, etcdRpcAddr)
	logger.Log.Info("网关监注册etcd完成")

	//初始化tcp服务
	mix_server.GMixServer.Init(etcdRpcAddr)

	//
	global_gateway.AllSessionMapInit()

	//监听grpc服务
	var tcpServerAddr, rpcServerAddr, httpServerAddr string
	tcpServerAddr = configx.GEnvCfg.GateWay.Dsn()
	rpcServerAddr = configx.GEnvCfg.GateWay.RpcDsn()
	httpServerAddr = configx.GEnvCfg.GateWay.HttpDsn()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runGrpcServer := func() {

		lisGrpc, err := net.Listen("tcp", rpcServerAddr)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("网关Server启动" + configx.GEnvCfg.GateWay.RpcDsn())
		mix_server.RunGrpcServer(ctx, lisGrpc, rpcServerAddr)

	}

	// 开始tcp服务
	runTcpServer := func() {

		lisTcp, err := net.Listen("tcp", tcpServerAddr)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("网关TcpServer启动" + configx.GEnvCfg.GateWay.Dsn())
		mix_server.RunTcpServer(ctx, lisTcp, etcdTcpAddr)
	}

	runHttpServer := func() {
		lisTcp, err := net.Listen("tcp", httpServerAddr)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("网关HttpServer启动" + configx.GEnvCfg.GateWay.HttpDsn())
		mix_server.RunHttpServer(ctx, lisTcp, httpServerAddr)
	}

	runGrpcServer()
	runTcpServer()
	runHttpServer()

	<-globals.SysSigChan
	log.Println("正在关闭 所有 服务...")
	cancel()

}
