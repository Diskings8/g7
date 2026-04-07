package model_game

import (
	"encoding/json"
	"fmt"
	"g7/common/logger"
	"g7/common/model_common"
	"time"
)

type Player struct {
	OnlineData // 存活数据
	AllBagData // 背包

	// 行为日志
	ActionLogs []*model_common.ActionLog

	// 在线数据
	IsOnline         bool      // 是否在线
	OfflineAt        time.Time // 离线时间
	OnlineAt         time.Time // 当前上线时间
	LastOfflineAt    time.Time // 上次离线时间
	LastDailyResetAt time.Time // 上次每日重置的时间
	LastWeekResetAt  time.Time // 每周重置时间
	LastMonthResetAt time.Time // 每月重置时间

	// 角色固定数据
	PlayerId int64
	UserId   int64
	ServerId int32
	Nickname string
}

func (p *Player) TableName() string {
	return "player"
}

func (this *Player) MarkOffLine() {
	this.IsOnline = false
	this.StreamConn = nil
	this.OfflineAt = time.Now()
}

func (this *Player) MarkOnline() {
	this.IsOnline = true
	// 在线时间在play_system更新
}

// RunInActor 发送逻辑到玩家玩家主协程（串行、安全）
func (p *Player) RunInActor(fn func()) {
	if !p.IsChanClosed {
		select {
		case p.ActionChan <- fn:
		default:
			logger.Log.Warn(fmt.Sprintf("player %d action is full", p.PlayerId))
		}
	}
}

// RunMainRoutine 玩家主协程处理核心逻辑和维护数据
func (p *Player) RunMainRoutine() {
	for !p.IsChanClosed {
		select {
		case fn, ok := <-p.ActionChan:
			if !ok {
				return
			}
			fn()
		}
	}
	//logger.Log.Info(fmt.Sprintf("main routine end: %d", p.PlayerId))
}

// RunSubRoutine 副协程处理数据落地
func (p *Player) RunSubRoutine() {
	for !p.IsChanClosed {
		select {
		case <-p.saveTicker.C:
			p.Save()
		case <-p.QuitChan:
			p.saveTicker.Stop()
			p.Save()
			p.IsChanClosed = true
		}
	}
	//logger.Log.Info(fmt.Sprintf("sub routine end: %d", p.PlayerId))
}

func (p *Player) Save() {
	//logger.Log.Info(fmt.Sprintf("save player %d", p.PlayerId))
	p.RunInActor(func() {
		// 1. 序列化（无竞争）
		dao := p.ToDao()
		// 2. 丢进全局DB队列（不阻塞）
		p.SaveChan <- dao
	})
}

func (p *Player) ToDao() *PlayerDao {
	dao := new(PlayerDao)
	dao.Nickname = p.Nickname
	dao.ServerId = p.ServerId
	dao.UserId = p.UserId
	dao.PlayerId = p.PlayerId
	dao.LastDailyResetAt = p.LastDailyResetAt
	dao.LastWeekResetAt = p.LastWeekResetAt
	dao.LastMonthResetAt = p.LastMonthResetAt

	// 通用数据
	var generalD = dao.GeneralD
	{
		generalD.BagData, _ = json.Marshal(p.Bags)
	}
	dao.generalData, _ = json.Marshal(generalD)

	// 养成数据
	var cultivationD = dao.CultivationD
	{

	}
	dao.cultivationData, _ = json.Marshal(cultivationD)
	// 活动数据
	var activityD = dao.ActivityD
	{

	}
	dao.activityData, _ = json.Marshal(activityD)
	return dao
}
