package mix_server

import (
	"g7/common/etcd"
	"g7/common/limiter"
	"g7/common/protos/pb"
)

var GMixServer = &GatewayMixServer{}

type GatewayMixServer struct {
	// common part
	etcdGrpcAddr string
	gameMonitor  *etcd.GameMonitor

	// grpc part
	pb.UnimplementedGatewayNodeServiceServer

	// tcp part
	// 限流
	ipLimiter         *limiter.IPLimiter         // 单 IP 限流
	connectionLimiter *limiter.ConnectionLimiter // 连接数限流
	rateLimiter       *limiter.RateLimiter
}

func (gms *GatewayMixServer) Init(etcdGrpcAddr string) {
	gms.etcdGrpcAddr = etcdGrpcAddr
	gms.gameMonitor = etcd.NewGameMonitor()
	//
	gms.ipLimiter = limiter.NewIPLimiter(100)
	gms.rateLimiter = limiter.NewRateLimiter(20000)
	gms.connectionLimiter = limiter.NewConnectionLimiter(5000)
	//
	gms.gameMonitor.Run()
}
