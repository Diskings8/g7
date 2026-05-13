package events

import "g7/common/protos/pb"

type Event struct {
	EventType int32
	SkillId   int32
	Seq       int32
	CasterId  int64
	Targets   []SkillUseResult
}

func (e *Event) ToProto() *pb.GirdEvents {
	result := &pb.GirdEvents{
		Seq:         e.Seq,
		EventType:   e.EventType,
		SkillId:     e.SkillId,
		CastActorId: e.CasterId,
	}
	for _, v := range e.Targets {
		result.Targets = append(result.Targets, v.ToProto())
	}
	return result
}

type SkillUseResult struct {
	TargetID    int64   `json:"targetId"`
	EffectScore float64 `json:"effectScore"` // 最终伤害（治疗为负数）
	IsCrit      bool    `json:"isCrit"`      // 是否暴击
	IsDodge     bool    `json:"isDodge"`     // 是否闪避
	IsBlock     bool    `json:"isBlock"`     // 是否格挡
}

func (sur *SkillUseResult) ToProto() *pb.SkillUseResult {
	return &pb.SkillUseResult{
		TargetId: sur.TargetID,
		Score:    sur.EffectScore,
		IsCrit:   sur.IsCrit,
		IsDodge:  sur.IsDodge,
		IsBlock:  sur.IsBlock,
	}
}
