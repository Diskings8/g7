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

func main() {
	// 1. 连接网关
	conn, err := net.Dial("tcp", "127.0.0.1:9001")
	gConn = conn
	if err != nil {
		fmt.Println("连接网关失败：", err)
		return
	}
	defer gConn.Close()

	fmt.Println("✅ 成功连接到 Gateway！")
	cronx.InitCron()
	firstMsg()
	//cronx.AddPer5SecondTask(heartbeat)

	MakeMsgToSend(pb.MsgID_MSG_Req_EnterGame, pb.Req_LoginGame{})

	// 等待接收网关返回
	//buf := make([]byte, 1024)
	for {
		msg, errx := protocol.ReadMessage(gConn)
		if errx == io.EOF {
			break
		}
		if msg == nil {
			continue
		}
		fmt.Printf("网关返回：MsgId:%d, %s", msg.MsgID, string(msg.Body))
		switch msg.MsgID {
		case pb.MsgID_MSG_Kick:
			_ = gConn.Close()
		}

	}
}

func MyData() pb.Req_AuthClientToGateWay {
	return pb.Req_AuthClientToGateWay{
		UerID:    2041160605846605824,
		PlayerID: 2041413406195585024,
		ServerID: 91001,
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyMDQxMTYwNjA1ODQ2NjA1ODI0LCJ1aWQiOjIwNDE0MTM0MDYxOTU1ODUwMjQsInNlcnZlcl9pZCI6OTEwMDEsImV4cCI6MTc3NTYzMjM5MSwiaWF0IjoxNzc1NTQ1OTkxfQ.fZxJNPdO0cy6odEUq7QUw6Rz-1AnGBAY9zpFoSsXxCc",
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
