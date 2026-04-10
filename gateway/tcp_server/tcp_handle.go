package tcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"g7/common/errcode"
	"g7/common/jwt"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/utils"
	"g7/gateway/global_gateway"
	"g7/gateway/tcp_session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"strings"
)

// 认证结构体

func (gts *GatewayTcpServer) HandleClient(conn net.Conn) {
	if ok, code, msg := gts.preCheck(conn); !ok {
		_ = protocol.WriteMessage(conn, 1002, errcode.MakeHttpErrCodeRespond(code, msg))
		_ = conn.Close()
		return
	}
	global_gateway.GCurrentConnection.Add(1)
	defer func() {
		_ = conn.Close()
		fmt.Println("connection closed")
		global_gateway.GCurrentConnection.Add(-1)
	}()

	sess := tcp_session.NewSession(conn)
	defer sess.Close()

	//log.Println("新连接:", conn.RemoteAddr())

	// 第一步：必须先认证（第一条消息）
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		return
	}

	if msg.MsgID != pb.MsgID_MSG_AUTH {
		log.Println("未认证，断开")
		return
	}

	// 解析认证
	var req pb.Req_AuthClientToGateWay
	_ = json.Unmarshal(msg.Body, &req)

	// 验证 Token（真实环境：调用登录服RPC/HTTP）
	if _, ok := checkToken(req.Token, req.GetUerID()); !ok {
		_ = protocol.WriteMessage(conn, 1002, errcode.MakeHttpErrCodeRespond(401, "token失效"))
		return
	}

	// --- 认证成功！会话赋值 ---
	sess.SetOwner(req.GetUerID(), req.GetPlayerID(), req.GetServerID())

	// --- 获取游戏服地址（从Watch缓存）---
	gameAddr, ok := gts.getGameServerAddr(utils.Int32ToString(req.ServerID))
	if !ok {
		_ = protocol.WriteMessage(conn, 1002, errcode.MakeHttpErrCodeRespond(503, "游戏服维护中"))
		return
	}

	// --- 连接游戏服 ---
	// 3. 连接游戏服，建立专属 gRPC 流
	stream, err := connectToGameServer(gameAddr)
	if err != nil {
		log.Printf("连接游戏服失败: %v", err)
		_ = protocol.WriteMessage(conn, 1002, errcode.MakeHttpErrCodeRespond(503, "连接游戏服失败"))
		return
	}
	sess.SetStream(stream)

	//log.Printf("认证成功：uid=%d role=%d serverID=%d", sess.userID, sess.playerID, sess.serverID)

	// --- 开始双向透传 ---
	go gatewayToClient(conn, sess, stream)
	clientToGateway(conn, sess, stream)
}

func (gts *GatewayTcpServer) preCheck(conn net.Conn) (bool, int, string) {
	clientIP := conn.RemoteAddr().String()
	// 截取 IP 部分（如果是IPv6或带端口）
	if idx := strings.Index(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx]
	}

	if !gts.connectionLimiter.Allow() {
		return false, 503, "系统繁忙"
	}

	if !gts.ipLimiter.Allow(clientIP) {
		return false, 429, "请求过于频繁"
	}

	if !gts.rateLimiter.Allow() {
		return false, 502, "服务器繁忙"
	}
	return true, 0, ""
}

func connectToGameServer(gameAddr string) (pb.GameStreamService_StreamClient, error) {
	conn, err := grpc.Dial(gameAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewGameStreamServiceClient(conn)
	stream, err := client.Stream(context.Background())
	return stream, err
}

// 客户端 → 网关 → 游戏服
func clientToGateway(conn net.Conn, _sess *tcp_session.Session, gameStream pb.GameStreamService_StreamClient) {
	makeFirstMessage(_sess, gameStream)
	for {
		msg, err := protocol.ReadMessage(conn)
		if err != nil {
			log.Printf("客户端断开: %v", err)
			return
		}
		// 把客户端的包转发给游戏服 gRPC 流
		_ = gameStream.Send(&pb.GameMessage{
			MsgId: uint32(msg.MsgID),
			Body:  msg.Body,
		})
	}
}

func makeFirstMessage(sess *tcp_session.Session, gameStream pb.GameStreamService_StreamClient) {
	msg := pb.Req_AuthClientToGame{PlayerID: sess.GetPlayerId()}
	msgBody, _ := json.Marshal(&msg)
	_ = gameStream.Send(&pb.GameMessage{
		MsgId: uint32(pb.MsgID_MSG_AUTH),
		Body:  msgBody,
	})
}

// 游戏服 → 网关 → 客户端
func gatewayToClient(conn net.Conn, sess *tcp_session.Session, gameStream pb.GameStreamService_StreamClient) {
	for {
		pkt, err := gameStream.Recv()
		if err != nil {
			log.Printf("游戏服流断开: %v", err)
			sess.Close()
			return
		}
		// 把游戏服的包转发给客户端 TCP
		_ = protocol.WriteMessage(conn, pb.MsgID(pkt.MsgId), pkt.Body)
	}
}

func checkToken(tokenStr string, clientUID int64) (*jwt.Claims, bool) {
	// 1. 本地直接解析校验
	claims, err := jwt.ParseToken(tokenStr)
	if err != nil {
		log.Printf("ParseToken error " + err.Error())
		return nil, false
	}

	// 2. 防篡改：客户端传的UID必须和Token里的UID一致
	if claims.UserID != clientUID {
		log.Printf("clientUID error " + fmt.Sprintf("%s, Req %s", claims.UID, clientUID))
		return nil, false
	}

	// 3. 校验成功！
	return claims, true
}
