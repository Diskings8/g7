package interface_game

import "g7/game/model_game"

type ISystem interface {
	Init()
	LoadData(PlayerDao *model_game.PlayerDao, Player *model_game.Player)
	OnEnterGame(Player *model_game.Player)
	GetName() string
}

type SaveSystem interface {
	SavePlayerDao(PlayerDao *model_game.SaveDaoD)
	GetName() string
}

type ResetSystem interface {
	DailyReset(Player *model_game.Player)
	WeekReset(Player *model_game.Player)
	MonthReset(Player *model_game.Player)
	GetName() string
}

type ActivitySystem interface {
	GetName() string
	Init()
	OnActivityOpen(cofId int32, activityId int64)
	OnActivityFinish(cofId int32, activityId int64)
	OnActivityClose(cofId int32, activityId int64)
}

type GoalSystem interface {
	Init()
	GetName() string
	OnGoalsUpdate([]int32)
	OnGoalsFinish([]int32)
}
