package interface_game

import "g7/game/model_game"

type ISystem interface {
	Init()
	LoadData(PlayerDao *model_game.PlayerDao, Player *model_game.Player)
	DailyReset(Player *model_game.Player)
	OnEnterGame(Player *model_game.Player)
}

type SaveSystem interface {
	SavePlayerDao(PlayerDao *model_game.PlayerDao)
	GetName() string
}
