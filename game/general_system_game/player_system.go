package general_system_game

import (
	"g7/common/globals"
	"g7/common/model_common"
	"g7/game/const_game"
	"g7/game/db_game"
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

func (this *playerSystem) SavePlayerDao(dao *model_game.PlayerDao) {
	//fmt.Printf("playerSystem save %d dao\n", dao.PlayerId)
	db_game.SetPlayerDao(dao)
}

func (this *playerSystem) GetName() string {
	return "playerSystem"
}

func (this *playerSystem) makeLoginLog(player *model_game.Player) {
	ld := model_common.ActionLog{
		BaseLog:      model_common.BaseLog{ServerId: player.ServerId, EventType: globals.ActionEventLogin, CreateTime: time.Now()},
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
