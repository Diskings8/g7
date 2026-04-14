package tcp_server

import (
	"g7/common/etcd"
	"g7/common/limiter"
)

var GTServer = &GatewayTcpServer{}

type GatewayTcpServer struct {
	gameMonitor *etcd.GameMonitor
	// 限流
	ipLimiter         *limiter.IPLimiter         // 单 IP 限流
	connectionLimiter *limiter.ConnectionLimiter // 连接数限流
	rateLimiter       *limiter.RateLimiter
}

func (gts *GatewayTcpServer) Init() {
	gts.gameMonitor = etcd.NewGameMonitor()
	//
	gts.ipLimiter = limiter.NewIPLimiter(100)
	gts.rateLimiter = limiter.NewRateLimiter(20000)
	gts.connectionLimiter = limiter.NewConnectionLimiter(5000)
	//
	gts.gameMonitor.Run()
}
