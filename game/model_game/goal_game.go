package model_game

type GoalManagerInterface interface {
	OnGoalsUpdate(map[int32][]int32)
	OnGoalsFinish(map[int32][]int32)
}

type GameGoal struct {
	SystemId    int32   // 注册的系统id
	Index       int32   // 系统内目标排序
	State       int32   // 目标进度
	GoalKind    int32   // 目标类型
	GoalObject  int32   // 目标对象
	Cnt         int64   // 当前值
	Requirement int64   // 所需值
	Params      []int64 // 其他参数
}

type GoalData struct {
	goals          map[int32][]*GameGoal
	callBackSystem GoalManagerInterface
}

func (gd *GoalData) Init() {
	gd.goals = make(map[int32][]*GameGoal)
}

func (gd *GoalData) SetCallBackSystem(cbs GoalManagerInterface) {
	gd.callBackSystem = cbs
}

func (gd *GoalData) GetCallBackSystem() GoalManagerInterface {
	return gd.callBackSystem
}

func (gd *GoalData) GetKindList(kind int32) []*GameGoal {
	return gd.goals[kind]
}

func (gd *GoalData) AddGoal(goal *GameGoal) {
	val, ok := gd.goals[goal.GoalKind]
	if !ok {
		gd.goals[goal.GoalKind] = append(gd.goals[goal.GoalKind], goal)
		return
	}
	for inx, v := range val {
		if v.SystemId == goal.SystemId && v.Index == goal.Index {
			val[inx] = goal
			return
		}
	}
	gd.goals[goal.GoalKind] = append(gd.goals[goal.GoalKind], goal)
}

func (gd *GoalData) ForgetGoal(goalKind, systemId, index int32) {
	val, ok := gd.goals[goalKind]
	if !ok {
		return
	}
	var removeIndex int
	for inx, v := range val {
		if v.SystemId == systemId && v.Index == index {
			removeIndex = inx
			break
		}
	}
	val = append(val[:removeIndex], val[removeIndex+1:]...)
	gd.goals[goalKind] = val
}
