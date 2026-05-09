package matchs

import (
	"fmt"
	"time"
)

type Matcher struct {
	pool      *LocalMatchPool
	matchType int32
	confId    int32
	cb        func(result *MatchResult)
}

func NewMatcher() *Matcher {
	return &Matcher{
		matchType: 1,
		confId:    1,
		pool:      NewLocalMatchPool(),
	}
}

func (mm *Matcher) GetCallbackFunc() func(result *MatchResult) {
	return mm.cb
}

func (mm *Matcher) SetCallbackFunc(cb func(result *MatchResult)) {
	mm.cb = cb
}

// 添加玩家到匹配队列
func (m *Matcher) Join(waiter WaitingInfo) error {
	//fmt.Println("new matcher team: ", waiter.TeamID)
	m.pool.Add(&waiter)
	return nil
}

// 取消匹配
func (m *Matcher) Cancel(waiterId string) {
	m.pool.Remove(waiterId)
}

// 尝试匹配（每次匹配尝试都调用）
func (m *Matcher) TryMatch() *MatchResult {
	// 1. 先处理扩圈
	m.processExpand()

	TeamCount := m.getTeamCountByMatchType()
	// 2. 获取所有等待玩家
	if m.pool.Size() < TeamCount {
		return nil // 不够10人
	}

	// 3. 遍历等待队列，尝试匹配
	// 这里简化：取第一个玩家作为锚点，尝试找9个队友
	waiters := m.pool.GetAllWaiters() // 需要实现这个方法
	for _, anchor := range waiters {
		result := m.tryMatchForWaiter(anchor)
		if result != nil {
			return result
		}
	}
	return nil
}

func (m *Matcher) getTeamCountByMatchType() int {
	switch m.matchType {
	case 1:
		return 1
	default:
		return competingTeamCount
	}
}

// 为指定玩家尝试匹配
func (m *Matcher) tryMatchForWaiter(anchor *WaitingInfo) *MatchResult {
	cfg := DefaultExpandConfig[anchor.ExpandLevel]
	minRating := anchor.Rating - cfg.Range
	maxRating := anchor.Rating + cfg.Range

	// 查找候选
	candidates := m.pool.FindByRatingRange(minRating, maxRating, anchor.TeamID, 50)
	waitCompetingTeamCount := m.getTeamCountByMatchType()
	if len(candidates) < (waitCompetingTeamCount - 1) {
		return nil // 候选不足
	}

	// 尝试组成队伍
	teamCount := m.getTeamCountByMatchType()
	teams := m.buildCompetingTeams(append([]WaitingItem{anchor.ToWaitingItem()}, candidates...), teamCount, 1)

	return &MatchResult{
		MatchID:  m.generateMatchID(),
		Teams:    teams,
		RoomType: m.matchType,
		ConfId:   m.confId,
	}
}

func (m *Matcher) buildCompetingTeams(items []WaitingItem, TeamCount, TeamSize int) []WaitingItem {
	return items[:TeamCount]
}

// 处理扩圈
func (m *Matcher) processExpand() {
	now := time.Now()
	needExpand := m.pool.GetPlayersNeedExpand(now)
	for _, waiter := range needExpand {
		newLevel := waiter.ExpandLevel + 1
		if newLevel < len(DefaultExpandConfig) {
			m.pool.UpdateExpandLevel(waiter.TeamID, newLevel)
		}
	}
}

func (m *Matcher) generateMatchID() string {
	switch m.confId {
	case 1:
		return fmt.Sprintf("match_1")
	default:
		return fmt.Sprintf("match_%d", time.Now().UnixNano())
	}
}
