package mix_server

import (
	"context"
	"fmt"
	"g7/common/jwt"
	"g7/common/netc"
	"g7/common/netc/tcp_conn"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/gateway/conn_session"
	"g7/gateway/global_gateway"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"strings"
)

func RunTcpServer(ctx context.Context, lis net.Listener, etcdTcpAddr string) {
	go func() {
		defer func() { recover() }()
		<-ctx.Done()
		conn_session.AllSessionClose()
		_ = lis.Close()
	}()

	go func() {
		defer func() { recover() }()
		for {
			conn, err := lis.Accept()
			if err != nil {
				return
			}
			go GMixServer.HandleClient(ctx, conn)
		}
	}()
	return
}

func (gms *GatewayMixServer) writeBackErrorMsg(conn netc.NetConnInterface, msg *pb.Rsp_AuthClientToGateWay) {
	byteData, _ := proto.Marshal(msg)
	conn.WriteToConn(0, &pb.GameMessage{MsgId: uint32(pb.MsgID_MSG_AUTH), Body: byteData})
}

func (gms *GatewayMixServer) HandleClient(ctx context.Context, nConn net.Conn) {
	conn := tcp_conn.NewNetConn(nConn)
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

	sess := conn_session.NewSession(ctx, conn, gms.etcdGrpcAddr)
	defer func() {
		gms.writeBackErrorMsg(conn, &pb.Rsp_AuthClientToGateWay{ErrCode: code, Reason: msg})
		sess.Close()
	}()

	//log.Println("新连接:", conn.RemoteAddr())

	// 第一步：必须先认证（第一条消息）
	packet, err := conn.ReadFromConn()
	if err != nil {
		code = 401
		msg = "请求失败"
		return
	}

	code, msg = gms.ServerAuth(sess, packet)
	if code != 0 {
		return
	}

}

func (gms *GatewayMixServer) preCheck(conn netc.NetConnInterface) (bool, int32, string) {
	clientIP := conn.RemoteAddr()
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
