package actors

import (
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/interfaces"
	"g7/comprehensive/model_compre/battle/skills"
	"g7/comprehensive/model_compre/battle/statemachine"
	"time"
)

type Vector3D struct {
	x float64
	y float64
	z float64
}

func (v3d *Vector3D) Add(add Vector3D) Vector3D {
	return *v3d
}

func (v3d *Vector3D) Mul(f float64) Vector3D {
	return *v3d
}

type Actor struct {
	id          int64
	actorType   int32
	pos         Vector3D
	velocity    Vector3D
	inputAction []*pb.GameMessage // 从房间 inputChan 收到的输入
	skills      []skills.Skill
	state       statemachine.StateMachine // 状态机管理待机、移动、攻击等

}

func (a *Actor) Init() {
}

func (a *Actor) ID() int64 {
	return a.id
}

func (a *Actor) Type() battle.ActorType {
	return battle.ActorType(a.actorType)
}

func (a *Actor) Update(delta time.Duration, world interfaces.World) {
	// 1. 处理所有缓存输入
	for _, input := range a.inputAction {
		a.processInput(input, world)
	}
	a.inputAction = a.inputAction[:0]

	// 2. 更新状态机（动画时间、技能冷却等）
	a.state.Update(delta)

	// 3. 应用移动
	a.pos = a.pos.Add(a.velocity.Mul(delta.Seconds()))
}

func (a *Actor) ToProto() any {
	return nil
}

func (a *Actor) AcceptInput(input actoractions.ActorAction) {
	a.inputAction = append(a.inputAction, input.Action)
}

func (a *Actor) processInput(action *pb.GameMessage, world interfaces.World) {

}
