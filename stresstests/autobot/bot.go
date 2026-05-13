package autobot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"g7/common/configx"
	"g7/common/cronx"
	"g7/common/dbc"
	"g7/common/dbc/dbc_interface"
	"g7/common/globals"
	"g7/common/model_common"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var GGameDb dbc_interface.DBInterface
var fileName = "./botPlayer.json"

func ConnectToDb() {
	var confStr = globals.GetEnvConfPath()
	configx.LoadEnvConf(confStr)
	GGameDb = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.Dsn())
	var l []model_common.GlobalPlayerIndex
	GGameDb.FindList(&l, map[string]any{"user_id": 2044258565992091648})
	//fmt.Println(len(l))

	url := "http://127.0.0.1:10081/api/player/select"

	ll := len(l)
	//ll := 1
	var rspData []*pb.Rsp_Http_SelectPlayer
	for i := 0; i < ll; i++ {
		var reqs = &pb.Req_Http_SelectPlayer{
			UID:      l[i].UserID,
			PlayerID: l[i].PlayerId,
			ServerID: l[i].ServerID,
		}
		d := doReq(url, reqs)
		if d == nil {
			continue
		}
		rspData = append(rspData, d)
	}
	saveToJSON(rspData, fileName)
}

var globalClient = &http.Client{
	Transport: &http.Transport{
		MaxConnsPerHost:     100, // 关键！最大并发连接
		MaxIdleConns:        100, // 关键！
		MaxIdleConnsPerHost: 100, // 关键！
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false, // 必须开启长连接
	},
	Timeout: 20 * time.Second,
}

func doReq(url string, reqData any) (rsp *pb.Rsp_Http_SelectPlayer) {
	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := globalClient.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close() // 必须关闭响应体
	bytesData, _ := io.ReadAll(resp.Body)
	rsp = &pb.Rsp_Http_SelectPlayer{}
	json.Unmarshal(bytesData, rsp)
	return
}

// saveToJSON 将记录列表保存到本地 JSON 文件
func saveToJSON(records []*pb.Rsp_Http_SelectPlayer, filename string) error {
	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 编码为 JSON 并写入（格式化输出，方便查看）
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

// loadFromJSON 从本地 JSON 文件加载记录
func loadFromJSON(filename string) ([]*pb.Rsp_Http_SelectPlayer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []*pb.Rsp_Http_SelectPlayer
	err = json.NewDecoder(file).Decode(&records)
	return records, err
}

var LocalAddr = "127.0.0.1:10001"

func LoadAndConn() {
	datas, err := loadFromJSON(fileName)
	if err != nil {
		return
	}
	testCount := 50
	testDuration := 3 * time.Minute

	lr := NewLatencyRecorder(testCount)
	cronx.InitCron()
	heartBeatChan := make(chan struct{}, 3)

	cronx.AddPer5SecondTask(func() { heartBeatChan <- struct{}{} })

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(testCount)

	for v := 0; v < testCount; v++ {
		go func(data *pb.Rsp_Http_SelectPlayer, index int) {
			defer wg.Done()
			randOneToConnect(ctx, data, lr, index, heartBeatChan)
		}(datas[v], v)
	}
	// 等所有玩家退出，或者10分钟超时
	fmt.Println("2 player in 1 room,", testDuration, " minute test for sum ", testCount, "player for ", testCount/2, " room ", time.Now())
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("所有玩家已退出，提前结束")
	case <-ctx.Done():
		fmt.Println(testDuration, "分钟压测时间到")
	}

	time.Sleep(1 * time.Second)
	fmt.Println(lr.Report())
}

func randOneToConnect(ctx context.Context, oneData *pb.Rsp_Http_SelectPlayer, lr *LatencyRecorder, index int, heartBeatChan chan struct{}) {
	var phase = &phaseState{State: 1}
	sendCh := make(chan *pb.GameMessage, 128)

	conn, err := net.Dial("tcp", LocalAddr)
	if err != nil {
		return
	}

	SendToConn := func() {
		for msg := range sendCh {
			_ = protocol.WritePacketToConn(conn, pb.MsgID(msg.MsgId), 0, msg.Body)
		}
	}

	msg := makeFirstMsg(oneData)
	makeMsgToSend(sendCh, pb.MsgID_MSG_AUTH, msg)
	makeMsgToSend(sendCh, pb.MsgID_MSG_Req_EnterGame, &pb.Req_LoginGame{})
	seqMap := make(map[int32]time.Time)
	seqLock := sync.RWMutex{}

	responseCh := make(chan int32, 100)
	var isConn = true

	heartbeatFunc := func() {
		for isConn {
			select {
			case <-heartBeatChan:
				heartbeat(conn)
			}
		}
	}

	waitChannel := func() {
		for {
			select {
			case c_Seq, _ := <-responseCh:
				seqLock.Lock()
				sendTime, ok := seqMap[c_Seq]
				if ok {
					delete(seqMap, c_Seq) // 读+删 一起锁
				}
				seqLock.Unlock()

				if !ok {
					fmt.Println(index, c_Seq, "not find in map")
					continue
				}

				// 计算延迟
				latency := time.Since(sendTime)
				lr.Record(latency)

			case <-ctx.Done():
				isConn = false
				conn.Close()
				return

			}
		}
	}

	runRecv := func() {
		for isConn {
			pkg, errx := protocol.ReadPacketFromConn(conn)
			if errx != nil {
				fmt.Println("网络断开")
				return
			}
			switchUnMarshal(oneData.PlayerID, pkg, responseCh, phase)
			pkg.Release()
		}
	}
	globalSeq := int32(1000 * index)

	runCount := 0
	skillCount := 0
	//
	runSend := func() {

		makeMsgToSend(sendCh, pb.MsgID_MSG_Req_StartMatch, &pb.Req_StartMatch{ConfId: 10})
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		moveCounter := 0
		bp := BotPlayer{}

		defer func() {
			fmt.Println(runCount, skillCount)
		}()
		for range ticker.C {
			select {
			case <-ctx.Done():
				return
			default:
				switch phase.State {
				case 1:
				case 2:
					phase.State = 3
					makeMsgToSend(sendCh, pb.MsgID_MSG_Req_EnterScene, &pb.Req_EnterRoom{})
					continue
				case 3:
					moveCounter++
					runCount++
					bp.randomWalk()
					makeMsgToSend(sendCh, pb.MsgID_MSG_Actor_Move, &pb.Action_Move{X: bp.X, Y: bp.Y})
					if moveCounter%2 == 0 {
						skillCount++
						seq := atomic.AddInt32(&globalSeq, 1)
						seqLock.Lock()
						seqMap[seq] = time.Now()
						seqLock.Unlock()
						makeMsgToSend(sendCh, pb.MsgID_MSG_Actor_UseSkill, &pb.Action_UseSkill{SkillId: int32(2010001), Seq: seq})
					}
				}

			}
		}
	}

	go SendToConn()
	go heartbeatFunc()
	go waitChannel()
	go runSend()
	runRecv()

}

func PacketSend(conn net.Conn) {

}

func makeMsgToSend(sendCh chan *pb.GameMessage, MsgId pb.MsgID, message proto.Message) (rsp any) {
	msgBody, _ := proto.Marshal(message)
	select {
	case sendCh <- &pb.GameMessage{
		MsgId:   uint32(MsgId),
		ErrCode: 0,
		Body:    msgBody,
	}:
	default:
		fmt.Println("send error")
	}
	return
}

func heartbeat(conn net.Conn) {
	_ = protocol.WritePacketToConn(conn, pb.MsgID_MSG_HeartBeat, 0, []byte(""))
	return
}

func makeFirstMsg(oneData *pb.Rsp_Http_SelectPlayer) *pb.Req_AuthClientToGateWay {
	firstMsg := &pb.Req_AuthClientToGateWay{
		UerID:    oneData.UserID,
		PlayerID: oneData.PlayerID,
		ServerID: oneData.ServerID,
		Token:    oneData.Token,
	}
	return firstMsg
}

func switchUnMarshal(playerId int64, pkg *protocol.Packet, c chan int32, phase *phaseState) {
	var data proto.Message
	var isSnaphot bool
	switch pkg.MsgID {
	case pb.MsgID_MSG_World_NineGridsSnapshot:
		data = &pb.WorldSnapshot{}
		isSnaphot = true
	case pb.MsgID_MSG_AUTH:
		data = &pb.Rsp_AuthClientToGateWay{}
	case pb.MsgID_MSG_Notify_MatchSuccess:
		data = &pb.Notify_MatchSuccess{}
		phase.State = 2
	default:
		return
	}
	err := proto.Unmarshal(pkg.Body, data)
	if err != nil {
		fmt.Printf("\n网关返回：MsgId:%d, %s\n", pkg.MsgID, err)
		return
	}
	if isSnaphot {
		rspD := data.(*pb.WorldSnapshot)
		//fmt.Printf("\nplayerId %d 网关返回：MsgId:%d, %+v\n", playerId, pkg.MsgID, rspD.Events)
		for _, v := range rspD.Events {
			if v.CastActorId == playerId {
				select {
				case c <- v.Seq:
				default:
					fmt.Println("responseCh 已满，丢弃 Seq:", v.Seq)
				}
			}
		}
	}
	//fmt.Printf("\n网关返回：MsgId:%d, %s\n", pkg.MsgID, data.String())
}
