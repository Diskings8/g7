package main

import (
	"flag"
	"g7/common/config"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/utils"
	"log"
	"net"
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

	// 3、注册etcd,监听游戏服
	etcd.InitETCD(config.GCfg.Etcd.Dsn)
	go etcd.WatchGameServers()
	etcd.RegisterGateway(config.GCfg.GateWay.Dsn())

	// 4、开始服务
	lis, err := net.Listen("tcp", config.GCfg.GateWay.Dsn())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("网关启动" + config.GCfg.GateWay.Dsn())

	for {
		conn, _ := lis.Accept()
		go HandleClient(conn)
		//go handle(conn)
	}
}

// 测试用
func handle(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	log.Println("客户端消息:", string(buf[:n]))

	// 示例：获取游戏服
	games, _ := etcd.GetGameServersByServerID("91001")
	if len(games) > 0 {
		log.Println("转发到游戏服:", games[0])
	}

	gateways, _ := etcd.GetAllGateways()
	if len(gateways) > 0 {
		log.Println("转发到游戏服:", gateways[0])
	}
}
