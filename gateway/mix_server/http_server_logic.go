package mix_server

import (
	"context"
	"g7/common/logger"
	"g7/common/netc/websocket_conn"
	"g7/gateway/conn_session"
	"g7/gateway/global_gateway"
	"log"
	"time"

	"net"
	"net/http"
)

func RunHttpServer(ctx context.Context, lis net.Listener, etcdHttpAddr string) {
	mux := http.NewServeMux()
	mux.Handle("/ws", GMixServer)

	server := &http.Server{
		Handler: mux,
	}
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Println("http server start error", err)
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdownCtx)
		}
	}()
	return
}

func (gms *GatewayMixServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ws":
		gms.upGradeWB(r.Context(), w, r)
	default:
		log.Println(r.URL.Path)
	}
}

func (gms *GatewayMixServer) upGradeWB(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	wbConn, err := gms.upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("升级失败：", err)
		return
	}
	global_gateway.GCurrentConnection.Add(1)
	defer func() {
		_ = wbConn.Close()
		//fmt.Println("connection closed")
		global_gateway.GCurrentConnection.Add(-1)
	}()

	conn := websocket_conn.NewWebSocketConn(wbConn)
	sess := conn_session.NewSession(ctx, conn, gms.etcdGrpcAddr)
	global_gateway.GConnSessionMap.NewSession(conn, sess)
	defer func() {
		sess.Close()
		conn.Close()
	}()
	// 第一步：必须先认证（第一条消息）
	packet, err := conn.ReadFromConn()

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	var code int32

	code, _ = gms.ServerAuth(sess, packet)
	if code != 0 {
		return
	}

}
