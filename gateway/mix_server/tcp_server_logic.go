package mix_server

import (
	"context"
	"fmt"
	"g7/common/jwt"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/utils"
	"g7/gateway/global_gateway"
	"g7/gateway/tcp_session"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"strings"
)

func (gms *GatewayMixServer) writeBackErrorMsg(conn net.Conn, msg *pb.Rsp_AuthClientToGateWay) {
	byteData, _ := proto.Marshal(msg)
	_ = protocol.WritePacketToConn(conn, pb.MsgID_MSG_AUTH, 0, byteData)
}

func (gms *GatewayMixServer) HandleClient(conn net.Conn) {
	var code int32
	var msg string
	var ok bool
	defer func() {
		gms.writeBackErrorMsg(conn, &pb.Rsp_AuthClientToGateWay{ErrCode: code, Reason: msg})
		_ = conn.Close()
	}()
	if ok, code, msg = gms.preCheck(conn); !ok {
		return
	}

	global_gateway.GCurrentConnection.Add(1)
	defer func() {
		_ = conn.Close()
		//fmt.Println("connection closed")
		global_gateway.GCurrentConnection.Add(-1)
	}()

	sess := tcp_session.NewSession(conn, gms.etcdGrpcAddr)
	defer func() {
		gms.writeBackErrorMsg(conn, &pb.Rsp_AuthClientToGateWay{ErrCode: code, Reason: msg})
		sess.Close()
	}()

	//log.Println("新连接:", conn.RemoteAddr())

	// 第一步：必须先认证（第一条消息）
	packet, err := protocol.ReadPacketFromConn(conn)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if packet.MsgID != pb.MsgID_MSG_AUTH {
		packet.Release()
		logger.Log.Info("未认证，断开")
		return
	}

	// 解析认证
	var req pb.Req_AuthClientToGateWay
	_ = proto.Unmarshal(packet.Body, &req)
	packet.Release()

	// 验证 Token（真实环境：调用登录服RPC/HTTP）
	if _, ok := gms.checkToken(req.Token, req.GetUerID()); !ok {
		code = 401
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
}

func (gms *GatewayMixServer) preCheck(conn net.Conn) (bool, int32, string) {
	clientIP := conn.RemoteAddr().String()
	// 截取 IP 部分（如果是IPv6或带端口）
	if idx := strings.Index(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx]
	}

	if !gms.connectionLimiter.Allow() {
		return false, 503, "系统繁忙"
	}

	if !gms.ipLimiter.Allow(clientIP) {
		return false, 429, "请求过于频繁"
	}

	if !gms.rateLimiter.Allow() {
		return false, 502, "服务器繁忙"
	}
	return true, 0, ""
}

func (gms *GatewayMixServer) checkToken(tokenStr string, clientUID int64) (*jwt.Claims, bool) {
	// 1. 本地直接解析校验
	claims, err := jwt.ParseToken(tokenStr)
	if err != nil {
		log.Printf("ParseToken error " + err.Error())
		return nil, false
	}

	// 2. 防篡改：客户端传的UID必须和Token里的UID一致
	if claims.UserID != clientUID {
		log.Printf("clientUID error " + fmt.Sprintf("%d, Req %d", claims.UID, clientUID))
		return nil, false
	}

	// 3. 校验成功！
	return claims, true
}

// 连接到游戏服
func (gms *GatewayMixServer) connectToGameServer(gameAddr string) (pb.GameStreamService_StreamClient, error) {

	client, err := protocol.NewGameNodeStreamClient(gameAddr)
	if err != nil {
		return nil, err
	}

	stream, err := client.Stream(context.Background())
	return stream, err
}

// 连接到room服
func (gms *GatewayMixServer) connectToRoomServer(gameAddr string) (pb.RoomStreamService_StreamClient, error) {

	client, err := protocol.NewRoomNodeStreamClient(gameAddr)
	if err != nil {
		return nil, err
	}

	stream, err := client.Stream(context.Background())
	return stream, err
}
