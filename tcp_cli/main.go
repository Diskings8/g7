package main

import (
	"encoding/json"
	"fmt"
	"g7/common/cronx"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"io"
	"net"
)

var gConn net.Conn

var RemoteAddr = "123.207.11.230:31001"
var LocalAddr = "127.0.0.1:10001"

func main() {
	// 1. 连接网关
	conn, err := net.Dial("tcp", LocalAddr)
	gConn = conn
	if err != nil {
		fmt.Println("连接网关失败：", err)
		return
	}
	defer gConn.Close()

	fmt.Println("✅ 成功连接到 Gateway！")
	cronx.InitCron()
	firstMsg()
	cronx.AddPer5SecondTask(heartbeat)

	MakeMsgToSend(pb.MsgID_MSG_Req_EnterGame, pb.Req_LoginGame{})

	// 等待接收网关返回
	//buf := make([]byte, 1024)
	for {
		msg, errx := protocol.ReadMessage(gConn)
		if errx == io.EOF {
			fmt.Println("网络断开")
			break
		}
		if msg == nil {
			continue
		}
		fmt.Printf("网关返回：MsgId:%d, %s\n", msg.MsgID, string(msg.Body))
		switch msg.MsgID {
		case pb.MsgID_MSG_Kick:
			_ = gConn.Close()
		}

	}
}

func MyData() pb.Req_AuthClientToGateWay {
	return pb.Req_AuthClientToGateWay{
		UerID:    2044258565992091648,
		PlayerID: 2044315259879165952,
		ServerID: 91001,
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyMDQ0MjU4NTY1OTkyMDkxNjQ4LCJ1aWQiOjIwNDQzMTUyNTk4NzkxNjU5NTIsInNlcnZlcl9pZCI6OTEwMDEsImV4cCI6MTc3NjMyNDExNCwiaWF0IjoxNzc2MjM3NzE0fQ.FkAh1ktHZe5sUAZhkihKFRQ-xTLpfVhjBtaUhk0gG0g",
	}
}

func firstMsg() {
	msg := MyData()
	msgBody, _ := json.Marshal(&msg)
	protocol.WriteMessage(gConn, pb.MsgID_MSG_AUTH, msgBody)
	return
}

func MakeMsgToSend(MsgId pb.MsgID, message any) {
	msgBody, _ := json.Marshal(&message)
	protocol.WriteMessage(gConn, MsgId, msgBody)
	return
}

func heartbeat() {
	protocol.WriteMessage(gConn, pb.MsgID_MSG_HeartBeat, []byte(""))
	return
}
