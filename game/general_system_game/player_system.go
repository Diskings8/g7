package general_system_game

import (
	"g7/game/const_game"
	"g7/game/db_game"
	"g7/game/manager_game"
	"g7/game/model_game"
)

var GPlayerSystem = &playerSystem{}

type playerSystem struct {
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_PlayerSystem, GPlayerSystem)
	manager_game.GSaveSystemManager.Register(const_game.General_PlayerSystem, GPlayerSystem)
}

func (this *playerSystem) Init() {

}

func (this *playerSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {

}

func (this *playerSystem) DailyReset(Player *model_game.Player) {}

func (this *playerSystem) OnEnterGame(Player *model_game.Player) {}

func (this *playerSystem) SavePlayerDao(dao *model_game.PlayerDao) {
	//fmt.Printf("playerSystem save %d dao\n", dao.PlayerId)
	db_game.SetPlayerDao(dao)
}

func (this *playerSystem) GetName() string {
	return "playerSystem"
}
