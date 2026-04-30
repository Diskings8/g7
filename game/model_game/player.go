package model_game

import (
	"fmt"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/protos/pb"
	"g7/common/utils"
	"time"
)

type Player struct {
	OnlineData  // 存活数据
	AllBagData  // 背包
	AllMailData // 邮件

	// 行为日志
	ActionLogs []*model_common.ActionLog

	// 在线数据
	IsDirty          bool      // 是否数据已被修改
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

// GetAllActionLogs 获取当前已执行的操作日志，在主协程执行
func (this *Player) GetAllActionLogs() []*model_common.ActionLog {
	var val = make([]*model_common.ActionLog, len(this.ActionLogs))
	copy(val, this.ActionLogs)
	this.ActionLogs = this.ActionLogs[:0]
	return val
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
		case <-p.QuitChan:
			return
		}
	}
	//logger.Log.Info(fmt.Sprintf("main routine end: %d", p.PlayerId))
}

// player.go 添加协程退出机制
func (p *Player) Close() {
	p.IsChanClosed = true
	close(p.ActionChan)
	p.QuitChan <- struct{}{}
}

func (p *Player) makeSaveDataAndSend(saveKind int) {
	dao := p.ToDao(saveKind)
	// 2. 丢进全局DB队列（不阻塞）
	p.SaveChan <- dao
}

func (p *Player) ToDao(kind int) *SaveDaoD {
	dao := new(PlayerDao)
	if kind == globals.SaveDataKindLoginOut {
		dao.OfflineAt = time.Now().Unix()
	}
	dao.Nickname = p.Nickname
	dao.ServerId = p.ServerId
	dao.UserId = p.UserId
	dao.PlayerId = p.PlayerId
	dao.LastDailyResetAt = p.LastDailyResetAt.Unix()
	dao.LastWeekResetAt = p.LastWeekResetAt.Unix()
	dao.LastMonthResetAt = p.LastMonthResetAt.Unix()
	dao.LastOfflineAt = p.LastOfflineAt.Unix()
	dao.OnlineAt = p.OnlineAt.Unix()

	// 通用数据
	var generalD = dao.GeneralD
	{
		generalD.BagData = p.AllBagData
		generalD.MailData = p.AllMailData
	}
	dao.GeneralData = utils.MarshalAndCompress(generalD)

	// 养成数据
	var cultivationD = dao.CultivationD
	{

	}
	dao.CultivationData = utils.MarshalAndCompress(cultivationD)
	// 活动数据
	var activityD = dao.ActivityD
	{

	}
	dao.ActivityData = utils.MarshalAndCompress(activityD)
	//
	daoD := &SaveDaoD{
		SaveType: kind,
		SaveData: dao,
	}
	return daoD
}

// Kick 必须在主线程执行
func (p *Player) Kick(reason string) {
	logger.Log.Debug(fmt.Sprintf("player %d kick success by reason:%s", p.PlayerId, reason))

	p.OfflineAt = time.Now()
	p.SendMessage(pb.MsgID_MSG_HeartBeat, &pb.Notify_Kick{Reason: reason})
	p.makeSaveDataAndSend(globals.SaveDataKindLoginOut)
	p.Close()
	if p.StreamCancelFunc != nil {
		p.StreamCancelFunc()
	}
	p.StreamConn = nil
}

func (p *Player) RedisReWrite(kind int) {
	//if !p.IsDirty {
	//	return
	//}
	p.makeSaveDataAndSend(kind)
}

func (p *Player) DbWrite(kind int) {
	p.makeSaveDataAndSend(kind)
}
