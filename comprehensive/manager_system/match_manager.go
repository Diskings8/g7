package manager_system

import (
	"fmt"
	"g7/comprehensive/model_compre/matchs"
	"sync"
	"time"
)

var GMatchManager = &matchManager{}

type matchManager struct {
	mu      sync.RWMutex
	matcher *matchs.Matcher
	cancel  chan struct{}
}

func (mm *matchManager) Init() {
	mm.matcher = matchs.NewMatcher()
	mm.cancel = make(chan struct{})
	mm.matcher.SetCallbackFunc(mm.callBackFunc)
	mm.start()
}

func (mm *matchManager) NewMatcher(playerId int64, serverId int32) error {
	matcherId := fmt.Sprintf("match_%d_%d", serverId, playerId)
	return mm.matcher.Join(matcherId, 1000, "", 0)
}

func (mm *matchManager) start() {
	go mm.matchLoop()
}

func (mm *matchManager) Stop() {
	mm.cancel <- struct{}{}
}

func (mm *matchManager) matchLoop() {
	ticker := time.NewTicker(1000 * time.Millisecond) // 每0.5秒尝试一次
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			result := mm.matcher.TryMatch()
			if result != nil {
				for _, pA := range result.TeamA {
					mm.matcher.Cancel(pA)
				}
				for _, pB := range result.TeamB {
					mm.matcher.Cancel(pB)
				}
				mm.onMatch(result)
			}
		case <-mm.cancel:
			return
		}
	}
}

func (mm *matchManager) onMatch(result *matchs.MatchResult) {
	mm.matcher.GetCallbackFunc()(result)
}

func (mm *matchManager) callBackFunc(result *matchs.MatchResult) {
	fmt.Printf("%+v\n", result)
	// 请求room 节点生成roomid

	// 广播
}
