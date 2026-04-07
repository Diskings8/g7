package interface_game

import "g7/game/model_game"

type ISystem interface {
	Init()
	LoadData(PlayerDao *model_game.PlayerDao, Player *model_game.Player)
	OnEnterGame(Player *model_game.Player)
	GetName() string
}

type SaveSystem interface {
	SavePlayerDao(PlayerDao *model_game.PlayerDao)
	GetName() string
}

type ResetSystem interface {
	DailyReset(Player *model_game.Player)
	WeekReset(Player *model_game.Player)
	MonthReset(Player *model_game.Player)
	GetName() string
}
