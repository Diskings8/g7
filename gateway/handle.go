package main

import (
	"context"
	"encoding/json"
	"fmt"
	"g7/common/errcode"
	"g7/common/etcd"
	"g7/common/jwt"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
)

// 认证结构体

func HandleClient(conn net.Conn) {
	defer conn.Close()
	sess := newSession(conn)
	defer sess.close()

	log.Println("新连接:", conn.RemoteAddr())

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
	sess.userID = req.GetUerID()
	sess.playerID = req.GetPlayerID()
	sess.serverID = req.ServerID

	// --- 获取游戏服地址（从Watch缓存）---
	gameAddr, ok := etcd.GetGameServerAddr(fmt.Sprintf("%d", req.ServerID))
	if !ok {
		_ = protocol.WriteMessage(conn, 1002, errcode.MakeHttpErrCodeRespond(503, "游戏服维护中"))
		return
	}

	// --- 连接游戏服 ---
	// 3. 连接游戏服，建立专属 gRPC 流
	stream, err := connectToGameServer(gameAddr)
	if err != nil {
		log.Printf("连接游戏服失败: %v", err)
		_ = protocol.WriteMessage(conn, 1002, []byte(`{"code":503,"msg":"连接游戏服失败"}`))
		return
	}
	sess.gameStream = stream

	log.Printf("认证成功：uid=%d role=%d serverID=%d", sess.userID, sess.playerID, sess.serverID)

	// --- 开始双向透传 ---
	go gatewayToClient(conn, sess, stream)
	clientToGateway(conn, sess, stream)
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
func clientToGateway(conn net.Conn, _sess *Session, gameStream pb.GameStreamService_StreamClient) {
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

func makeFirstMessage(sess *Session, gameStream pb.GameStreamService_StreamClient) {
	msg := pb.Req_AuthClientToGame{PlayerID: sess.playerID}
	msgBody, _ := json.Marshal(&msg)
	_ = gameStream.Send(&pb.GameMessage{
		MsgId: uint32(pb.MsgID_MSG_AUTH),
		Body:  msgBody,
	})
}

// 游戏服 → 网关 → 客户端
func gatewayToClient(conn net.Conn, sess *Session, gameStream pb.GameStreamService_StreamClient) {
	for {
		pkt, err := gameStream.Recv()
		if err != nil {
			log.Printf("游戏服流断开: %v", err)
			sess.close()
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
