package world

import (
	"g7/comprehensive/model_compre/battle/actoractions"
	"time"
)

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
		}
	}

	// 角色执行操作
	for _, a := range w.actors {
		a.Update(delta, w)
	}
	w.mu.RUnlock()

	//todo other event
}
