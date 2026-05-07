package actors

import (
	"g7/comprehensive/model_compre/battle/world"
	"time"
)

type PlayerActor struct {
	Actor
}

func (p *PlayerActor) HandleInput(input *any, world *world.World) {
}

func (p *PlayerActor) Update(delta time.Duration, world *world.World) {

}
