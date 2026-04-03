package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"net"
)

func main() {
	// 1. 连接网关
	conn, err := net.Dial("tcp", "127.0.0.1:9001")
	if err != nil {
		fmt.Println("连接网关失败：", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ 成功连接到 Gateway！")

	// 发送消息到网关
	//msg := "hello gateway!!"
	//_, err = conn.Write([]byte(msg))
	//if err != nil {
	//	fmt.Println("发送失败：", err)
	//	return
	//}
	firstMsg(conn)

	MakeMsgToSend(conn, pb.MsgID_MSG_ENTER_GAME, pb.Req_LoginGame{})

	// 等待接收网关返回
	buf := make([]byte, 1024)
	n, _ := bufio.NewReader(conn).Read(buf)
	fmt.Println("网关返回：", string(buf[:n]))
}

func MyData() pb.Req_AuthClientToGateWay {
	return pb.Req_AuthClientToGateWay{
		UID:      910001,
		RoleID:   10241,
		ServerID: 91001,
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMDI0MSwidWlkIjo5MTAwMDEsInNlcnZlcl9pZCI6OTEwMDEsImV4cCI6MTc3NTI4NjYxMSwiaWF0IjoxNzc1MjAwMjExfQ.P_tqKk4QGzPsxfXF5wtvFvog777OYGrc5ii1poIBeLM",
	}
}

func firstMsg(conn net.Conn) {
	msg := MyData()
	msgBody, _ := json.Marshal(&msg)
	protocol.WriteMessage(conn, pb.MsgID_MSG_AUTH, msgBody)
	return
}

func MakeMsgToSend(conn net.Conn, MsgId pb.MsgID, message any) {
	msgBody, _ := json.Marshal(&message)
	protocol.WriteMessage(conn, MsgId, msgBody)
	return
}
