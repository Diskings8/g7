package grids

import "g7/comprehensive/model_compre/battle/common_battle"

const GridSize = 20.0 // 每格边长

type GirdMap struct {
	girdMapData map[[2]int32][]int64
}

func (gm *GirdMap) Init() {
	gm.girdMapData = make(map[[2]int32][]int64)
}

func (gm *GirdMap) Clear() {
	clear(gm.girdMapData)
}

func (gm *GirdMap) GridCoord(pos common_battle.Vector3D) (x, y int32) {
	return int32(pos.X / GridSize), int32(pos.Y / GridSize)
}

func (gm *GirdMap) SetKeyActor(key [2]int32, actorId int64) {
	gm.girdMapData[key] = append(gm.girdMapData[key], actorId)
}

func (gm *GirdMap) GetKeyActor(key [2]int32) []int64 {
	return gm.girdMapData[key]
}

func (gm *GirdMap) NineGrids(x, y int32) [][2]int32 {
	return [][2]int32{
		{x - 1, y - 1}, {x, y - 1}, {x + 1, y - 1},
		{x - 1, y}, {x, y}, {x + 1, y},
		{x - 1, y + 1}, {x, y + 1}, {x + 1, y + 1},
	}
}
