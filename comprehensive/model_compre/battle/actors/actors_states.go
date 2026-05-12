package actors

import "g7/comprehensive/model_compre/battle"

type Attributes struct {
	Hp    float64
	MaxHp float64
	Mp    float64
	MaxMp float64

	MoveSpeed float64
}

func (as *Attributes) DefaultAttributes() {
	as.MaxHp = 100
	as.Hp = as.MaxHp
	as.MaxMp = 100
	as.Mp = as.MaxHp

	as.MoveSpeed = 5
}

func (as *Attributes) IsEnough(attri int, val float64) bool {
	switch attri {
	case battle.AttributesHp:
		return as.Hp >= val
	case battle.AttributesMp:
		return as.Mp >= val
	default:
		return false
	}
}

func (as *Attributes) Cost(attri int, val float64) {
	switch attri {
	case battle.AttributesHp:
		as.Hp = max(as.Mp-val, 0)
		as.Hp = min(as.Hp, as.MaxHp)
	case battle.AttributesMp:
		as.Mp = max(as.MaxMp-val, 0)
	default:

	}
}

type States struct {
	IsDead   bool
	IsMoving bool
}

func (s *States) DefaultStates() {

}
