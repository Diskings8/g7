package service_login

import (
	"g7/common/etcd"
	"g7/common/limiter"
)

var LTServer = &loginHttpServer{}

type loginHttpServer struct {
	GameMonitor *etcd.GameMonitor

	// 限流
	ipLimiter         *limiter.IPLimiter         // 单 IP 限流
	connectionLimiter *limiter.ConnectionLimiter // 连接数限流
	rateLimiter       *limiter.RateLimiter
}

func (hts *loginHttpServer) Init() {
	hts.GameMonitor = etcd.NewGameMonitor()

	//
	hts.ipLimiter = limiter.NewIPLimiter(100)
	hts.rateLimiter = limiter.NewRateLimiter(20000)
	hts.connectionLimiter = limiter.NewConnectionLimiter(5000)

	//
	hts.GameMonitor.Run()
}
