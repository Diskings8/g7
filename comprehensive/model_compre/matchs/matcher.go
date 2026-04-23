package matchs

import (
	"fmt"
	"time"
)

type Matcher struct {
	pool *LocalMatchPool
	cb   func(result *MatchResult)
}

func NewMatcher() *Matcher {
	return &Matcher{
		pool: NewLocalMatchPool(),
	}
}

func (mm *Matcher) GetCallbackFunc() func(result *MatchResult) {
	return mm.cb
}

func (mm *Matcher) SetCallbackFunc(cb func(result *MatchResult)) {
	mm.cb = cb
}

// 添加玩家到匹配队列
func (m *Matcher) Join(matcher string, rating int, teamID string, teamSize int) error {
	fmt.Println("new matcher", matcher)
	waiter := &WaitingInfo{
		PlayerID:    matcher,
		Rating:      rating,
		EnterTime:   time.Now(),
		ExpandLevel: 0,
		TeamID:      teamID,
		TeamSize:    teamSize,
	}
	m.pool.Add(waiter)
	return nil
}

// 取消匹配
func (m *Matcher) Cancel(playerID string) {
	m.pool.Remove(playerID)
}

// 尝试匹配（每次匹配尝试都调用）
func (m *Matcher) TryMatch() *MatchResult {
	// 1. 先处理扩圈
	m.processExpand()

	// 2. 获取所有等待玩家
	if m.pool.Size() < TeamCount {
		return nil // 不够10人
	}

	// 3. 遍历等待队列，尝试匹配
	// 这里简化：取第一个玩家作为锚点，尝试找9个队友
	players := m.pool.GetAllWaiters() // 需要实现这个方法
	for _, anchor := range players {
		result := m.tryMatchForWaiter(anchor)
		if result != nil {
			return result
		}
	}
	return nil
}

// 为指定玩家尝试匹配
func (m *Matcher) tryMatchForWaiter(anchor *WaitingInfo) *MatchResult {
	cfg := DefaultExpandConfig[anchor.ExpandLevel]
	minRating := anchor.Rating - cfg.Range
	maxRating := anchor.Rating + cfg.Range

	// 查找候选
	candidates := m.pool.FindByRatingRange(minRating, maxRating, anchor.PlayerID, 50)
	if len(candidates) < (TeamCount - 1) {
		return nil // 候选不足
	}

	// 尝试组成10人队伍
	team := m.buildBalancedTeam(append([]string{anchor.PlayerID}, candidates...))
	if len(team) != TeamCount {
		return nil
	}

	// 分成两队
	teamA, teamB := m.splitTeams(team)

	return &MatchResult{
		RoomID:  generateRoomID(),
		TeamA:   teamA,
		TeamB:   teamB,
		RatingA: m.calcTeamRating(teamA),
		RatingB: m.calcTeamRating(teamB),
	}
}

// 构建平衡队伍（简单的贪心算法）
func (m *Matcher) buildBalancedTeam(candidates []string) []string {
	if len(candidates) < TeamCount {
		return nil
	}
	// 简化：取前10个
	return candidates[:TeamCount]
}

// 分成两队（尽量平衡）
func (m *Matcher) splitTeams(players []string) ([]string, []string) {
	// 按分数排序后，蛇形分配
	// 简化：前5后5
	return players[:TeamCount/2], players[TeamCount/2:]
}

// 计算队伍平均分
func (m *Matcher) calcTeamRating(team []string) int {
	total := 0
	for _, pid := range team {
		if waiter := m.pool.Get(pid); waiter != nil {
			total += waiter.Rating
		}
	}
	if len(team) == 0 {
		return 0
	}
	return total / len(team)
}

// 处理扩圈
func (m *Matcher) processExpand() {
	now := time.Now()
	needExpand := m.pool.GetPlayersNeedExpand(now)
	for _, waiter := range needExpand {
		newLevel := waiter.ExpandLevel + 1
		if newLevel < len(DefaultExpandConfig) {
			m.pool.UpdateExpandLevel(waiter.PlayerID, newLevel)
		}
	}
}

func generateRoomID() string {
	return fmt.Sprintf("room_%d", time.Now().UnixNano())
}
