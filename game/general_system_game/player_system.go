package general_system_game

import (
	"g7/common/globals"
	"g7/common/model_common"
	"g7/common/mqc/mq_topic"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/db_game"
	"g7/game/global_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"time"
)

var GPlayerSystem = &playerSystem{}

type playerSystem struct {
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_PlayerSystem, GPlayerSystem)
	manager_game.GSaveSystemManager.Register(const_game.General_PlayerSystem, GPlayerSystem)
	manager_game.GResetSystemManager.Register(const_game.General_PlayerSystem, GPlayerSystem)
}

func (this *playerSystem) Init() {

}

func (this *playerSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {
	Player.LastDailyResetAt = utils.FormatTimestamp(dao.LastMonthResetAt)
	Player.LastWeekResetAt = utils.FormatTimestamp(dao.LastWeekResetAt)
	Player.LastMonthResetAt = utils.FormatTimestamp(dao.LastMonthResetAt)
}

func (this *playerSystem) DailyReset(Player *model_game.Player) {
	Player.LastDailyResetAt = time.Now()
}

func (this *playerSystem) WeekReset(Player *model_game.Player) {
	Player.LastWeekResetAt = time.Now()
}

func (this *playerSystem) MonthReset(Player *model_game.Player) {
	Player.LastMonthResetAt = time.Now()
}

func (this *playerSystem) OnEnterGame(Player *model_game.Player) {
	Player.IsOnline = true
	Player.OnlineAt = time.Now()
	Player.LastOfflineAt = Player.OfflineAt
	Player.OfflineAt = time.Time{}
	this.makeLoginLog(Player)
}

func (this *playerSystem) SavePlayerDao(daoD *model_game.SaveDaoD) {
	//fmt.Printf("playerSystem save %d dao\n", dao.PlayerId)
	//logger.Log.Info(fmt.Sprintf("save player dao %#v", *daoD))
	switch daoD.SaveType {
	case globals.SaveDataKindCornCache:
		_ = global_game.GPlayerCache.SetPlayerCache(daoD.SaveData)
	case globals.SaveDataKindCornDb:
		db_game.SetPlayerDao(daoD.SaveData)
	case globals.SaveDataKindLoginOut:
		_ = global_game.GPlayerCache.SetPlayerCache(daoD.SaveData)
		db_game.SetPlayerDao(daoD.SaveData)
		this.makeOffLineLog(daoD.SaveData)

	}
}

func (this *playerSystem) GetName() string {
	return "playerSystem"
}

func (this *playerSystem) makeLoginLog(player *model_game.Player) {
	ld := model_common.ActionLog{
		BaseLog:      model_common.BaseLog{ServerId: player.ServerId, EventType: globals.ActionEventLogin, CreateTime: time.Now().Unix()},
		PlayerID:     player.PlayerId,
		Action:       "Login",
		Reason:       "",
		CostItem:     nil,
		CostCurrency: nil,
		GainItem:     nil,
		GainCurrency: nil,
		Ext:          "",
	}
	player.ActionLogs = append(player.ActionLogs, &ld)
}

func (this *playerSystem) makeOffLineLog(dao *model_game.PlayerDao) {
	ld := model_common.ActionLog{
		BaseLog:      model_common.BaseLog{ServerId: dao.ServerId, EventType: globals.ActionEventLogout, CreateTime: time.Now().Unix()},
		PlayerID:     dao.PlayerId,
		Action:       "Logout",
		Reason:       "",
		CostItem:     nil,
		CostCurrency: nil,
		GainItem:     nil,
		GainCurrency: nil,
		Ext:          "",
	}
	global_game.GGlobalMQ.ProduceMessage(mq_topic.MakeGameActionTopicKey(), ld)
}
