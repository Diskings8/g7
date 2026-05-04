package general_system_game

import (
	"fmt"
	"g7/common/logger"
	"g7/game/const_game"
	"g7/game/manager_game"
	"g7/game/model_game"
)

var GGoalSystem = &goalSystem{}

type goalSystem struct {
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_GoalSystem, GGoalSystem)
	manager_game.GGoalSystemManager.Register(const_game.General_GoalSystem, GGoalSystem)
}

func (this *goalSystem) Init() {
}

func (this *goalSystem) GetName() string {
	return "general_bag_system"
}

func (this *goalSystem) LoadData(_ *model_game.PlayerDao, Player *model_game.Player) {
	Player.GoalData.Init()
}

func (this *goalSystem) DailyReset(Player *model_game.Player) {}

func (this *goalSystem) OnEnterGame(Player *model_game.Player) {

}

func (this *goalSystem) OnGoalsUpdate(goals []int32) {
	logger.Log.Info(fmt.Sprintf("%+v", goals))
}

func (this *goalSystem) OnGoalsFinish(goals []int32) {
	logger.Log.Info(fmt.Sprintf("%+v", goals))
}

func (this *goalSystem) AddGoal(goal *model_game.GameGoal, player *model_game.Player) {
	player.GoalData.AddGoal(goal)
}

func (this *goalSystem) DevNewGoal(i int32) (g *model_game.GameGoal) {
	switch i {
	case 1:
		g = &model_game.GameGoal{
			SystemId:    const_game.General_GoalSystem,
			Index:       1,
			State:       model_game.GoalRunning,
			GoalKind:    const_game.GoalType_KillMonster,
			GoalObject:  102,
			Cnt:         0,
			Requirement: 3,
			Params:      nil,
		}
	default:
		g = &model_game.GameGoal{
			SystemId:    const_game.General_GoalSystem,
			Index:       2,
			State:       model_game.GoalRunning,
			GoalKind:    const_game.GoalType_LevelUp,
			GoalObject:  0,
			Cnt:         0,
			Requirement: 3,
			Params:      nil,
		}

	}
	return
}
