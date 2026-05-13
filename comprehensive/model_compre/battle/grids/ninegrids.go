package grids

import (
	"g7/common/utils"
	"g7/comprehensive/model_compre/battle/common_battle"
	"g7/comprehensive/model_compre/battle/events"
	"sync"
)

const (
	MapSize     = 1000 // 地图总边长 1000
	HalfMapSize = 500  // 半张地图 500（坐标范围 -500 ~ 500）
	GridCount   = 20   // 固定 20x20 格子
	GridSize    = 50   // 每个格子大小 = 1000 / 20 = 50
)

type OneGridData struct {
	ActorIds []int64
	Events   []events.Event
}

type GridMap struct {
	rw          sync.RWMutex
	gridMapData [GridCount][GridCount]*OneGridData
	dirtyx      []int32
	dirtyy      []int32
}

func (gm *GridMap) Init() {
	gm.gridMapData = [GridCount][GridCount]*OneGridData{}
	for k1 := range gm.gridMapData {
		for k2 := range gm.gridMapData[k1] {
			gm.gridMapData[k1][k2] = &OneGridData{}
		}
	}
}

func (gm *GridMap) GridEventClear() {
	for inx := range gm.dirtyx {
		row := gm.dirtyx[inx]
		col := gm.dirtyy[inx]
		gm.gridMapData[row][col].Events = gm.gridMapData[row][col].Events[:0]
	}
}

func (gm *GridMap) GridCoord(pos common_battle.Vector3D) (int32, int32) {
	offsetX := pos.X + HalfMapSize // -500~500 → 0~1000
	offsetY := pos.Y + HalfMapSize // -500~500 → 0~1000

	// 边界保护（防止坐标越界）
	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}
	if offsetX >= MapSize {
		offsetX = MapSize - 1
	}
	if offsetY >= MapSize {
		offsetY = MapSize - 1
	}

	// 第二步：转成 20x20 格子编号
	col := int32(offsetX) / GridSize
	row := int32(offsetY) / GridSize

	// 最终一定是 0~19
	return row, col
}

func (gm *GridMap) checkKey(x, y int32) bool {
	if x < 0 || y < 0 || x > 19 || y > 19 {
		return true
	}
	return false
}

func (gm *GridMap) SetGridActor(x, y int32, actorId int64) {
	if gm.checkKey(x, y) {
		return
	}
	gm.rw.Lock()
	defer gm.rw.Unlock()
	oneGrid := gm.gridMapData[x][y]
	if oneGrid == nil {
		oneGrid = &OneGridData{}
	}
	oneGrid.ActorIds = append(oneGrid.ActorIds, actorId)
	gm.gridMapData[x][y] = oneGrid
}

func (gm *GridMap) SetGridEvent(x, y int32, event events.Event) {
	if gm.checkKey(x, y) {
		return
	}
	gm.rw.Lock()
	defer gm.rw.Unlock()
	oneGrid := gm.gridMapData[x][y]
	if oneGrid == nil {
		oneGrid = &OneGridData{}
	}
	oneGrid.Events = append(oneGrid.Events, event)
	gm.gridMapData[x][y] = oneGrid
	gm.dirtyx = append(gm.dirtyx, x)
	gm.dirtyy = append(gm.dirtyy, y)
}

func (gm *GridMap) GetOneGrid(x, y int32) *OneGridData {
	if gm.checkKey(x, y) {
		return nil
	}
	return gm.gridMapData[x][y]
}

func (gm *GridMap) GetGridActors(x, y int32) []int64 {
	if gm.checkKey(x, y) {
		return nil
	}
	oneGrid := gm.gridMapData[x][y]
	if oneGrid != nil {
		return oneGrid.ActorIds
	}
	return nil
}

func (gm *GridMap) RemoveActor(x, y int32, actorId int64) {
	if gm.checkKey(x, y) {
		return
	}
	gm.rw.Lock()
	defer gm.rw.Unlock()
	oneGrid := gm.gridMapData[x][y]
	utils.RemoveAllElement(oneGrid.ActorIds, actorId)
}

func (gm *GridMap) GetGridEvent(x, y int32) []events.Event {
	if gm.checkKey(x, y) {
		return nil
	}
	oneGrid := gm.gridMapData[x][y]
	if oneGrid != nil {
		return oneGrid.Events
	}
	return nil
}

func (gm *GridMap) NineGrids(x, y int32) [][2]int32 {
	return [][2]int32{
		{x - 1, y - 1}, {x, y - 1}, {x + 1, y - 1},
		{x - 1, y}, {x, y}, {x + 1, y},
		{x - 1, y + 1}, {x, y + 1}, {x + 1, y + 1},
	}
}
