package matchs

import (
	"time"
)

var TeamCount = 2

// 扩圈配置
type ExpandConfig struct {
	Level   int // 扩圈等级 0,1,2,3,4
	Range   int // 分数范围 ±
	MaxWait int // 最大等待秒数
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
	PlayerID    string
	Rating      int       // 隐藏分
	EnterTime   time.Time // 进入匹配时间
	ExpandLevel int       // 当前扩圈等级
	TeamID      string    // 队伍ID（组队时）
	TeamSize    int       // 队伍人数
}

// 匹配结果
type MatchResult struct {
	RoomID  string   `json:"room_id"`
	TeamA   []string `json:"team_a"`
	TeamB   []string `json:"team_b"`
	RatingA int      `json:"rating_a"`
	RatingB int      `json:"rating_b"`
}
