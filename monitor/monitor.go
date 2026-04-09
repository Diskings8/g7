package main

import (
	"context"
	"fmt"
	"g7/common/configx"
	"g7/common/globals"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"g7/common/etcd"
	"google.golang.org/grpc"
)

// -------------------------- 全局配置 --------------------------
const (
	AlarmServiceName        = "game-server"
	GrpcSlowThreshold       = 500 * time.Millisecond // gRPC慢请求阈值
	ConnCountAlarmThreshold = 5000                   // 连接数超警戒值
	RedisMemAlarmThreshold  = 800                    // Redis内存MB警戒值
	KafkaLagAlarmThreshold  = 1000                   // Kafka堆积条数阈值
)

// -------------------------- 启动所有监控 --------------------------
func StartAllMonitor() {
	go WatchServiceHealth()   // 服务宕机监控
	go WatchConnectionCount() // 连接数监控
	go WatchRedisMemory()     // Redis内存监控
	go WatchKafkaLag()        // Kafka堆积监控

	log.Println("✅ 所有监控启动完成")
}

// -------------------------- 1. 服务宕机监控（ETCD） --------------------------
func WatchServiceHealth() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var (
		lastGameAlarmTime    time.Time
		lastGatewayAlarmTime time.Time
		lastLoginAlarmTime   time.Time
		alarmInterval        = 5 * time.Minute // 5分钟内只发一次告警
	)

	for range ticker.C {
		gameList := etcd.GetServiceList(globals.GameServer)
		gatewayList := etcd.GetServiceList(globals.GateWayServer)
		loginList := etcd.GetServiceList(globals.LoginServer)

		now := time.Now()
		if len(gameList) == 0 && now.Sub(lastGameAlarmTime) > alarmInterval {
			SendAlarm("【严重】游戏服全部宕机！")
			lastGameAlarmTime = now
		}
		if len(gatewayList) == 0 && now.Sub(lastGatewayAlarmTime) > alarmInterval {
			SendAlarm("【严重】网关全部宕机！")
			lastGatewayAlarmTime = now
		}
		if len(loginList) == 0 && now.Sub(lastLoginAlarmTime) > alarmInterval {
			SendAlarm("【严重】登录服全部宕机！")
			lastLoginAlarmTime = now
		}
	}
}

// -------------------------- 2. gRPC 耗时监控拦截器 --------------------------
func GrpcMonitorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		start := time.Now()
		resp, err := handler(ctx, req)
		cost := time.Since(start)

		if cost > GrpcSlowThreshold {
			SendAlarm(fmt.Sprintf("【性能】gRPC慢请求 %s 耗时:%dms",
				info.FullMethod, cost.Milliseconds()))
		}
		return resp, err
	}
}

// -------------------------- 3. 连接数监控 --------------------------
func WatchConnectionCount() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		count := GetCurrentConnCount()
		if count > ConnCountAlarmThreshold {
			SendAlarm(fmt.Sprintf("【预警】连接数过高 当前:%d", count))
		}
	}
}

// GetCurrentConnCount获取当前网关连接数
func GetCurrentConnCount() int32 {
	var sum int32
	for _, v := range etcd.GetServiceList(globals.GateWayServer) {

		c, _ := protocol.NewGatewayNodeClient(context.Background(), v)
		rps, _ := c.GetConnCount(context.Background(), &pb.Req_Node_ConnCount{})
		sum += rps.GetCount()
	}
	return sum
}

// -------------------------- 4. Redis 内存监控 --------------------------
func WatchRedisMemory() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		usedMB := redisx.GetUsedMemoryMB()
		if usedMB > RedisMemAlarmThreshold {
			SendAlarm(fmt.Sprintf("【预警】Redis内存过高 当前:%dMB", usedMB))
		}
	}
}

// -------------------------- 5. Kafka 堆积监控 --------------------------
func WatchKafkaLag() {
	//ticker := time.NewTicker(10 * time.Second)
	//defer ticker.Stop()
	//
	//for range ticker.C {
	//	lag := kafka.GetConsumerLag("game-log-group")
	//	if lag > KafkaLagAlarmThreshold {
	//		SendAlarm(fmt.Sprintf("【预警】Kafka堆积 当前:%d条", lag))
	//	}
	//}
}

// -------------------------- 统一告警入口 --------------------------
func SendAlarm(msg string) {
	alarmMsg := fmt.Sprintf("[%s] %s", AlarmServiceName, msg)
	log.Println("🔴 " + alarmMsg)

	// 对接钉钉/企业微信机器人
	// go SendDingAlarm(alarmMsg)
}

func main() {
	confStr := globals.ConfDev
	configx.LoadEnvConf(confStr)
	//
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	// 初始化redis
	redisx.Init(configx.GEnvCfg.Redis.Addr, configx.GEnvCfg.Redis.Password, configx.GEnvCfg.Redis.DB)
	//

	StartAllMonitor()
	log.Println("✅ 监控服务启动成功")

	// 3. 优雅退出（比 select {} 更健壮）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("🛑 监控服务退出")
}
