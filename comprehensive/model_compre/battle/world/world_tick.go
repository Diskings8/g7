package world

import (
	"fmt"
	"g7/comprehensive/model_compre/battle/actoractions"
	"time"
)

func (w *World) rebuildGrid() {
	w.gridMap.Clear()
	w.mu.RLock()
	defer w.mu.RUnlock()
	for id, a := range w.actors {
		gx, gy := w.gridMap.GridCoord(a.Pos())
		key := [2]int32{gx, gy}
		w.gridMap.SetKeyActor(key, id)
	}
}

func (w *World) Tick(delta time.Duration) {
	w.frameID++
	w.eventLog = w.eventLog[:0] // 清空事件

	// 1. 收集本帧所有输入（非阻塞，一次清空 channel）
	var actions []actoractions.ActorAction
	for {
		select {
		case act := <-w.actionsCh:
			actions = append(actions, act)
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
		} else {
			fmt.Println("act", act.ActorId)
		}
	}

	// 角色执行操作
	for _, a := range w.actors {
		a.Update(delta, w)
	}
	w.mu.RUnlock()

	//todo other event

	w.rebuildGrid()
}
