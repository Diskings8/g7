package main

import (
	"bufio"
	"fmt"
	"g7/common/cronx"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"os"
)

var gConn net.Conn

var RemoteAddr = "123.207.11.230:31001"
var LocalAddr = "127.0.0.1:10001"
var c chan struct{}

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

	MakeMsgToSend(pb.MsgID_MSG_Req_EnterGame, &pb.Req_LoginGame{})

	// 等待接收网关返回
	//buf := make([]byte, 1024)
	go WaitWrite()
	for {
		msg, errx := protocol.ReadMessage(gConn)
		if errx == io.EOF {
			fmt.Println("网络断开")
			break
		}
		if msg == nil {
			continue
		}
		fmt.Printf("\n网关返回：MsgId:%d, %s\n", msg.MsgID, string(msg.Body))
		switch msg.MsgID {
		case pb.MsgID_MSG_Kick:
			_ = gConn.Close()
		}
		c <- struct{}{}
	}
}

func MyData() pb.Req_AuthClientToGateWay {
	return pb.Req_AuthClientToGateWay{
		UerID:    2044258565992091648,
		PlayerID: 2044315259879165952,
		ServerID: 91001,
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyMDQ0MjU4NTY1OTkyMDkxNjQ4LCJ1aWQiOjIwNDQzMTUyNTk4NzkxNjU5NTIsInNlcnZlcl9pZCI6OTEwMDEsImV4cCI6MTc3NjM5MzA1NSwiaWF0IjoxNzc2MzA2NjU1fQ._erIPy-e59QohQMm1hjxzUPrG0lEqOP1CcrqwHnTRhQ",
	}
}

func firstMsg() {
	msg := MyData()
	msgBody, _ := proto.Marshal(&msg)
	protocol.WriteMessage(gConn, pb.MsgID_MSG_AUTH, msgBody)
	return
}

func MakeMsgToSend(MsgId pb.MsgID, message proto.Message) (rsp any) {
	msgBody, _ := proto.Marshal(message)
	protocol.WriteMessage(gConn, MsgId, msgBody)
	return
}

func heartbeat() {
	protocol.WriteMessage(gConn, pb.MsgID_MSG_HeartBeat, []byte(""))
	return
}

func WaitWrite() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n输入 指令：")
		scanner.Scan()
		key := scanner.Text()
		fmt.Println(key)
		MakeMsgToSend(pb.MsgID_MSG_GM_Cmd, &pb.Req_RunGm{Cmd: key})
	}
}
