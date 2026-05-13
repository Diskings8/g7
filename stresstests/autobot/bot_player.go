package autobot

import (
	"math"
	"math/rand"
)

const (
	MapMinX = -500
	MapMaxX = 500
	MapMinY = -500
	MapMaxY = 500
)

type BotPlayer struct {
	X float64
	Y float64
	Z float64
}

var MoveStepPerSecond = 5.0

func (vp *BotPlayer) randomWalk() {
	// 随机方向，每秒移动一个固定的步长
	angle := rand.Float64() * 2 * math.Pi
	dx := math.Cos(angle) * MoveStepPerSecond
	dy := math.Sin(angle) * MoveStepPerSecond

	vp.X += dx
	vp.Y += dy

	// 边界钳位
	if vp.X < MapMinX {
		vp.X = MapMinX
	}
	if vp.X > MapMaxX {
		vp.X = MapMaxX
	}
	if vp.Y < MapMinY {
		vp.Y = MapMinY
	}
	if vp.Y > MapMaxY {
		vp.Y = MapMaxY
	}
}
