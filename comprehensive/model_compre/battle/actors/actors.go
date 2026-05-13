package actors

import (
	"fmt"
	"g7/common/confs"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/comprehensive/model_compre/battle"
	"g7/comprehensive/model_compre/battle/actoractions"
	"g7/comprehensive/model_compre/battle/common_battle"
	"g7/comprehensive/model_compre/battle/events"
	"g7/comprehensive/model_compre/battle/interfaces"
	"g7/comprehensive/model_compre/battle/skills"
	"g7/comprehensive/model_compre/battle/statemachine"
	"github.com/golang/protobuf/proto"
	"math"
	"time"
)

func NewActor(actorId int64, actorType battle.ActorType, pos common_battle.Vector3D, world interfaces.World) Actor {
	a := Actor{
		id:        actorId,
		actorType: actorType,
	}
	a.CurPos = pos
	a.Init(world)
	return a
}

type Actor struct {
	Attributes
	States
	Skills
	common_battle.Pos
	KillerId  int64
	dirty     bool
	id        int64
	world     interfaces.World
	actorType battle.ActorType

	inputAction []*pb.GameMessage         // 从房间 inputChan 收到的输入
	state       statemachine.StateMachine // 状态机管理待机、移动、攻击等
}

var _ interfaces.Actor = (*Actor)(nil)

func (a *Actor) Init(world interfaces.World) {
	a.world = world
	a.Attributes.DefaultAttributes()
	a.States.DefaultStates()
	a.skills = make(map[int32]skills.Skill)
	a.skillCDs = make(map[int32]time.Duration)

}

func (a *Actor) ID() int64 {
	return a.id
}

func (a *Actor) Type() battle.ActorType {
	return a.actorType
}

func (a *Actor) IsDirty() bool {
	return a.dirty
}

func (a *Actor) IsStateMoving() bool {
	return a.States.IsMoving
}

func (a *Actor) ClearDirty() {
	a.dirty = false
}

func (a *Actor) Update(delta time.Duration, world interfaces.World) {
	// 1. 处理所有缓存输入
	for _, input := range a.inputAction {
		a.processInput(input, world)
	}
	a.inputAction = a.inputAction[:0]

	// 2. 执行 3D 移动
	a.doMove(delta)

	// 3. 状态机更新
	a.state.Update(delta)

	// 4.cd更新
	a.Skills.UpdateAllCD(delta)
}

func (a *Actor) AcceptInput(input actoractions.ActorAction) {
	a.inputAction = append(a.inputAction, input.Action)
}

func (a *Actor) processInput(action *pb.GameMessage, world interfaces.World) {
	a.dirty = true
	switch pb.MsgID(action.GetMsgId()) {
	case pb.MsgID_MSG_Actor_Move:
		var moveReq pb.Action_Move
		_ = proto.Unmarshal(action.Body, &moveReq)
		a.TargetPos = common_battle.Vector3D{
			X: moveReq.X,
			Y: moveReq.Y,
			Z: moveReq.Z, // 大部分游戏 Y 是高度
		}
		a.States.IsMoving = true
	case pb.MsgID_MSG_Actor_UseSkill:
		var useSkillReq pb.Action_UseSkill
		err := proto.Unmarshal(action.Body, &useSkillReq)
		if err != nil {
			logger.Log.Warn(fmt.Sprintf("%+v:%s", err.Error(), string(action.Body)))
			return
		}
		a.castSkill(&useSkillReq, world)
	}
}

func (a *Actor) doMove(delta time.Duration) {
	if !a.States.IsMoving {
		return
	}
	// 帧时间（秒）
	dt := delta.Seconds()

	// 3D 方向向量（你的坐标规则）
	a.OldPos = a.CurPos
	dx := a.TargetPos.X - a.CurPos.X // 前后
	dy := a.TargetPos.Y - a.CurPos.Y // 左右
	dz := a.TargetPos.Z - a.CurPos.Z // 上下

	// 3D 距离平方（优化，不用开根号）
	distSq := dx*dx + dy*dy + dz*dz
	if distSq < 0.0001 {
		a.CurPos = a.TargetPos
		a.States.IsMoving = false
		return
	}
	// 3D 距离
	dist := math.Sqrt(distSq)

	// 单位方向向量
	dirX := dx / dist
	dirY := dy / dist
	dirZ := dz / dist

	// 每帧移动步长
	step := a.MoveSpeed * dt

	// 3D 位置更新（完全匹配你的坐标系）
	a.CurPos.X += dirX * step
	a.CurPos.Y += dirY * step
	a.CurPos.Z += dirZ * step
}

func (a *Actor) GetPos() common_battle.Pos {
	return a.Pos
}

func (a *Actor) castSkill(reqSkill *pb.Action_UseSkill, world interfaces.World) {
	if a.States.IsDead {
		return
	}
	skillId := reqSkill.GetSkillId()
	targetIds := reqSkill.GetTargetIds()
	conf, ok := confs.GConfigSkill.Find(skillId)
	if !ok {
		logger.Log.Warn(fmt.Sprintf("%d skill %d not exist", a.id, skillId))
		return
	}
	if !a.Skills.IsCDOk(skillId) {
		logger.Log.Warn(fmt.Sprintf("%d skill %d not cd enough", a.id, skillId))
		return
	}
	if !a.Attributes.IsEnough(battle.AttributesMp, conf.Skillcost) {
		logger.Log.Warn(fmt.Sprintf("%d skill %d not cost %f enough", a.id, skillId, conf.Skillcost))
		return
	}
	a.Attributes.Cost(battle.AttributesMp, conf.Skillcost)

	a.Skills.StartSkillCD(skillId, time.Duration(conf.Cd))
	// 判断对象合法性

	targets := world.FindActors(a, targetIds, conf.Range)
	if targets == nil {
		logger.Log.Warn(fmt.Sprintf("%d skill %d not range actors ", a.id, skillId))
		return
	}
	var effectScore float64
	switch conf.Skilltype {
	case 1:
		effectScore = conf.Score * (-1)
	default:
		effectScore = conf.Score
	}
	var OneEvent = events.Event{
		CasterId: a.id,
		SkillId:  skillId,
		Seq:      reqSkill.GetSeq(),
	}
	for _, v := range targets {
		v.TakeEffect(a.id, effectScore)
		OneEvent.Targets = append(OneEvent.Targets, events.SkillUseResult{
			TargetID:    v.ID(),
			EffectScore: effectScore,
			IsCrit:      false,
			IsDodge:     false,
			IsBlock:     false,
		})
	}
	world.AddEvent(a.CurPos, OneEvent)
	return
}
