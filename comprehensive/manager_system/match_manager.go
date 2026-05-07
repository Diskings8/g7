package manager_system

import (
	"context"
	"fmt"
	"g7/common/etcd"
	"g7/common/logger"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/utils"
	"g7/comprehensive/model_compre/matchs"
	"strings"
	"sync"
	"time"
)

var GMatchManager = &matchManager{}

type matchManager struct {
	mu          sync.RWMutex
	confMatcher map[int32]*matchs.Matcher
	cancel      chan struct{}
	confKeys    []int32
}

func (mm *matchManager) Init() {
	mm.confMatcher = make(map[int32]*matchs.Matcher, 0)
	mm.cancel = make(chan struct{})
	mm.start()
}

func (mm *matchManager) NewMatcher(matcher *pb.MatchWaiter) matchs.WaitingInfo {
	return matchs.WaitingInfo{
		ConfId:      1,
		TeamID:      matcher.GetTeamId(),
		Rating:      matcher.GetScore(),
		EnterTime:   time.Now(),
		ExpandLevel: 0,
		TeamSize:    len(matcher.GetTeamMember()),
		TeamLeader:  matcher.TeamLeader,
		TeamMember:  matcher.GetTeamMember(),
	}
}

func (mm *matchManager) getConfMatcher(confId int32) *matchs.Matcher {
	mm.mu.RLock()
	val, ok := mm.confMatcher[confId]
	mm.mu.RUnlock()
	if !ok {
		mm.mu.Lock()
		mm.confMatcher[confId] = matchs.NewMatcher()
		mm.confMatcher[confId].SetCallbackFunc(mm.callBackFunc)
		mm.confKeys = append(mm.confKeys, confId)
		mm.mu.Unlock()
		val = mm.confMatcher[confId]
	}
	return val
}

func (mm *matchManager) getAllConfKeys() []int32 {
	return mm.confKeys
}

func (mm *matchManager) StarMatch(matcher matchs.WaitingInfo) error {
	return mm.getConfMatcher(matcher.ConfId).Join(matcher)
}

func (mm *matchManager) start() {
	go mm.matchLoop()
}

func (mm *matchManager) Stop() {
	mm.cancel <- struct{}{}
}

func (mm *matchManager) matchLoop() {
	ticker := time.NewTicker(1000 * time.Millisecond) // 每1秒尝试一次
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mm.allConfTryMatch()
		case <-mm.cancel:
			return
		}
	}
}

func (mm *matchManager) allConfTryMatch() {
	for _, key := range mm.getAllConfKeys() {
		result := mm.getConfMatcher(key).TryMatch()
		if result != nil {
			for _, pT := range result.Teams {
				mm.getConfMatcher(key).Cancel(pT.TeamID)
			}
			mm.onMatch(result)
		}
	}
}

func (mm *matchManager) onMatch(result *matchs.MatchResult) {
	// 需要异步处理分发
	go mm.getConfMatcher(result.ConfId).GetCallbackFunc()(result)
}

func (mm *matchManager) callBackFunc(result *matchs.MatchResult) {
	// 提取队伍人员
	var memberLoginKeys = make([]string, 0, 6)
	for _, AOne := range result.Teams {
		for _, member := range AOne.TeamMembers {
			sL := strings.Split(member, "_")
			if len(sL) < 2 {
				continue
			}
			memberLoginKeys = append(memberLoginKeys, redisx.MakePlayerLoginKey(utils.StringToInt32(sL[0]), utils.StringToInt64(sL[1])))
		}
	}
	//创建房间
	createRoomRsp := mm.authorityCreateRoom(result, memberLoginKeys)
	logger.Log.Info(fmt.Sprintf("authorityCreateRoom:resp:%+v", createRoomRsp))

	var serverMap = make(map[string][]string)
	var allMemberLoginVals = make([]string, 0, 6)
	memberLoginVals, err := redisx.MGet(memberLoginKeys)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, oneLoginVal := range memberLoginVals {
		loginVal := oneLoginVal.(string)
		gameServerAddr := redisx.GetPlayerLoginValueIndex(loginVal, redisx.RedisPlayerLoginValIndexGameServerAddr)
		serverMap[gameServerAddr] = append(serverMap[gameServerAddr], loginVal)
		allMemberLoginVals = append(allMemberLoginVals, loginVal)
	}

	ctxReq, cancelReq := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelReq()

	for k, v := range serverMap {
		client, errNode := protocol.NewGameNodeClient(k)
		if errNode != nil {
			logger.Log.Warn(errNode.Error() + k)
			continue
		}
		nodeReq := &pb.Req_Node_MatchSuccess{
			State:       1,
			MatchMember: v,
			RoomId:      createRoomRsp.GetRoomId(),
			RoomAddr:    createRoomRsp.GetRoomAddr(),
		}
		client.MatchNodeMatchSuccess(ctxReq, nodeReq)
	}
}

func (mm *matchManager) authorityCreateRoom(result *matchs.MatchResult, members []string) (rsp *pb.Rsp_Node_CreateRoom) {
	kvL, err := etcd.GetRoomManagerServersList()
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	if len(kvL) == 0 {
		logger.Log.Error("not GetRoomManagerServersList")
		return
	}
	severAddr := kvL[0].V
	cli, err := protocol.NewRoomManagerNodeClient(severAddr)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	rsp, err = cli.AuthorityCreateRoom(context.Background(),
		&pb.Req_Node_CreateRoom{RoomType: result.RoomType, MatchMember: members,
			ConfId: result.ConfId, MatchId: result.MatchID})
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	return
}
