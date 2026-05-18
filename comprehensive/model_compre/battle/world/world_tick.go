package world

import (
	"fmt"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/interfaces"
	"time"
)

func (w *World) Tick(delta time.Duration) {
	w.frameID++
	w.gridMap.GridEventClear() // 清空事件

	// 1. 收集本帧所有输入（非阻塞，一次清空 channel）
	var actions []actoractions.ActorAction
	for {
		select {
		case act := <-w.actionsCh:
			actions = append(actions, act)
			//fmt.Println("action:", actions)
		default:
			goto DONE
		}
	}
DONE:
	// 赋予角色操作
	w.mu.RLock()
	for _, act := range actions {
		if a, ok := w.actors[act.ActorId]; ok {
			a.AcceptInput(act)
		}
	}

	// 角色执行操作
	for _, a := range w.actors {
		a.Update(delta, w)
		w.UpdateActorGrid(a)
	}
	w.mu.RUnlock()

	//todo other event
}

func (w *World) UpdateActorGrid(a interfaces.Actor) {
	if a.IsStateMoving() {
		newx, newy := w.gridMap.GridCoord(a.GetPos().CurPos)
		oldx, oldy := w.gridMap.GridCoord(a.GetPos().OldPos)

		if newx != oldx || newy != oldy {
			w.gridMap.RemoveActor(oldx, oldy, a.ID())
			w.gridMap.SetGridActor(newx, newy, a.ID())
			fmt.Println("change gird", newx, newy)
		}
	}
}

func (w *World) InitActorGrid(a interfaces.Actor) {
	newx, newy := w.gridMap.GridCoord(a.GetPos().CurPos)
	w.gridMap.SetGridActor(newx, newy, a.ID())
}
