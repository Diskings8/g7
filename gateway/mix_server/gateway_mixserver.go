package mix_server

import (
	"fmt"
	"g7/common/etcd"
	"g7/common/limiter"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/utils"
	"g7/gateway/conn_session"
	"github.com/golang/protobuf/proto"
	"net/http"

	"github.com/gorilla/websocket"
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

	// http part
	upGrader websocket.Upgrader
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

	gms.upGrader = websocket.Upgrader{
		// 允许跨域（本地测试/前端联调必须开）
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func (gms *GatewayMixServer) ServerAuth(sess *conn_session.Session, packet *protocol.Packet) (code int32, msg string) {
	if packet.MsgID != pb.MsgID_MSG_AUTH {
		packet.Release()
		logger.Log.Info("未认证，断开")
		code = 402
		msg = "未认证token失效"
		return
	}

	// 解析认证
	var req pb.Req_AuthClientToGateWay
	_ = proto.Unmarshal(packet.Body, &req)
	packet.Release()

	// 验证 Token（真实环境：调用登录服RPC/HTTP）
	if _, ok := gms.checkToken(req.Token, req.GetUerID()); !ok {
		code = 402
		msg = "token失效"
		return
	}

	// --- 认证成功！会话赋值 ---
	sess.SetOwner(req.GetUerID(), req.GetPlayerID(), req.GetServerID())
	// --- 获取游戏服地址（从Watch缓存）---
	gameAddr, ok := gms.gameMonitor.GetGameServerAddr(utils.Int32ToString(req.ServerID), utils.Int64ToString(req.GetPlayerID()))
	if !ok {
		code = 503
		msg = "游戏服维护中"
		return
	}
	// --- 连接游戏服 ---
	// 3. 连接游戏服，建立专属 gRPC 流
	stream, err := gms.connectToGameServer(gameAddr)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("连接游戏服失败: %v", err))
		code = 503
		msg = "连接游戏服失败"
		return
	}
	sess.SetGameStream(stream)

	//log.Printf("认证成功：uid=%d role=%d serverID=%d", sess.userID, sess.playerID, sess.serverID)

	// --- 开始双向透传 ---
	go sess.RunGoRoutineToRecvFromGame()
	go sess.RunGoRoutineToSendToGame()
	go sess.RunGoRoutineToSendToConn()
	sess.RunGoRoutineToRecvFromConn()
	return
}
