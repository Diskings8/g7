package model_game

const (
	GoalRunning = int32(iota) + 1
	GoalFinish
)

type Trigger struct {
	p *Player
}

func NewTrigger(p *Player) Trigger {
	return Trigger{p: p}
}

func (t *Trigger) TriggerCommon(EventType, Object int32, cnt int64) {
	list := t.p.GoalData.GetKindList(EventType)
	if len(list) == 0 {
		return
	}
	var updateSystem = make(map[int32][]int32)
	var finishSystem = make(map[int32][]int32)
	for index, v := range list {
		if v.State != GoalRunning {
			continue
		}
		if v.GoalObject == Object {
			list[index].Cnt += cnt
			if list[index].Cnt >= list[index].Requirement {
				list[index].State = GoalFinish
				finishSystem[v.SystemId] = append(finishSystem[v.SystemId], v.Index)
				continue
			}
			updateSystem[v.SystemId] = append(updateSystem[v.SystemId], v.Index)
		}
	}
	t.p.GoalData.GetCallBackSystem().OnGoalsUpdate(updateSystem)
	t.p.GoalData.GetCallBackSystem().OnGoalsFinish(finishSystem)
}
