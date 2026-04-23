package main

import (
	"bufio"
	"flag"
	"fmt"
	"g7/common/cronx"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/utils"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"os"
	"time"
)

var gConn net.Conn

var RemoteAddr = "123.207.11.230:31001"
var LocalAddr = "127.0.0.1:10001"
var c chan struct{}

var selectRole int32
var cmdParms string

func main() {
	flag.StringVar(&cmdParms, "role", "1", "")
	flag.Parse()
	selectRole = utils.StringToInit32(cmdParms)
	// 1. 连接网关
	go runConnect()
	time.Sleep(1 * time.Second)
	fmt.Println("✅ 成功连接到 Gateway！:", SelectUse(selectRole).PlayerID)
	cronx.InitCron()
	firstMsg()
	cronx.AddPer5SecondTask(heartbeat)

	MakeMsgToSend(pb.MsgID_MSG_Req_EnterGame, &pb.Req_LoginGame{})

	// 等待接收网关返回
	//buf := make([]byte, 1024)
	WaitWrite()
}

func runConnect() {
	if gConn == nil {
		conn, err := net.Dial("tcp", LocalAddr)
		gConn = conn
		if err != nil {
			fmt.Println("连接网关失败：", err)
			return
		}

	}
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
	}
	gConn.Close()
	gConn = nil
	time.Sleep(3 * time.Second)
	runConnect()
}

func MyData() *pb.Req_AuthClientToGateWay {
	return SelectUse(selectRole)
}

func SelectUse(int int32) *pb.Req_AuthClientToGateWay {
	switch int {
	case 1:
		return &pb.Req_AuthClientToGateWay{
			UerID:    2044258565992091648,
			PlayerID: 2044315259879165952,
			ServerID: 91001,
			Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyMDQ0MjU4NTY1OTkyMDkxNjQ4LCJ1aWQiOjIwNDQzMTUyNTk4NzkxNjU5NTIsInNlcnZlcl9pZCI6OTEwMDEsImV4cCI6MTc3OTUzMDE5NSwiaWF0IjoxNzc2OTM4MTk1fQ.j-UwFzARER7RF02dBvEoOr_4TlB3dfAml6lvDIjmW3M",
		}
	case 2:
		return &pb.Req_AuthClientToGateWay{
			UerID:    2044258565992091648,
			PlayerID: 2045039918497009664,
			ServerID: 91001,
			Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyMDQ0MjU4NTY1OTkyMDkxNjQ4LCJ1aWQiOjIwNDUwMzk5MTg0OTcwMDk2NjQsInNlcnZlcl9pZCI6OTEwMDEsImV4cCI6MTc3OTUzMDI0NiwiaWF0IjoxNzc2OTM4MjQ2fQ.43iQRR_Qcb8Okwm2resEjstzP7NTK8xp8iKUX40a7TA",
		}
	}
	return nil
}

func firstMsg() {
	msg := MyData()
	msgBody, _ := proto.Marshal(msg)
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
