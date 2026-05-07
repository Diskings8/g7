package matchs

import (
	"time"
)

var competingTeamCount = 2
var TeamSize = 1

// 扩圈配置
type ExpandConfig struct {
	Level   int     // 扩圈等级 0,1,2,3,4
	Range   float64 // 分数范围 ±
	MaxWait int     // 最大等待秒数
}

var DefaultExpandConfig = []ExpandConfig{
	{Level: 0, Range: 100, MaxWait: 10},
	{Level: 1, Range: 200, MaxWait: 20},
	{Level: 2, Range: 400, MaxWait: 40},
	{Level: 3, Range: 800, MaxWait: 80},
	{Level: 4, Range: 99999, MaxWait: 999},
}

// 等待玩家信息
type WaitingInfo struct {
	TeamID      string  // 队伍ID
	Rating      float64 // 隐藏分
	ConfId      int32
	EnterTime   time.Time // 进入匹配时间
	ExpandLevel int       // 当前扩圈等级
	TeamSize    int       // 队伍人数
	TeamLeader  string    // 队长
	TeamMember  []string  // 队员
}

func (wi WaitingInfo) ToWaitingItem() WaitingItem {
	return WaitingItem{
		Score:       wi.Rating,
		TeamMembers: wi.TeamMember,
		TeamID:      wi.TeamID,
	}
}

type WaitingItem struct {
	TeamID      string // 队伍ID
	TeamMembers []string
	Score       float64
}

// 匹配结果
type MatchResult struct {
	MatchID  string `json:"match_id"`
	ConfId   int32
	RoomType int32
	Teams    []WaitingItem `json:"teams"`
}
