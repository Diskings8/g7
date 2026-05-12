package grids

import (
	"g7/comprehensive/model_compre/battle/common_battle"
	"g7/comprehensive/model_compre/battle/events"
)

const GridSize = 20.0 // 每格边长

type oneGirdData struct {
	ActorIds []int64
	Events   []events.Event
}

type GirdMap struct {
	girdMapData map[[2]int32]*oneGirdData
}

func (gm *GirdMap) Init() {
	gm.girdMapData = make(map[[2]int32]*oneGirdData)
}

func (gm *GirdMap) GirdActorsClear() {
	for k := range gm.girdMapData {
		gm.girdMapData[k].ActorIds = gm.girdMapData[k].ActorIds[:0]
	}
}

func (gm *GirdMap) GirdEventClear() {
	for k := range gm.girdMapData {
		gm.girdMapData[k].Events = gm.girdMapData[k].Events[:0]
	}
}

func (gm *GirdMap) GridCoord(pos common_battle.Vector3D) (x, y int32) {
	return int32(pos.X / GridSize), int32(pos.Y / GridSize)
}

func (gm *GirdMap) SetGirdActor(key [2]int32, actorId int64) {
	oneGird, ok := gm.girdMapData[key]
	if !ok {
		oneGird = &oneGirdData{}
	}
	oneGird.ActorIds = append(oneGird.ActorIds, actorId)
	gm.girdMapData[key] = oneGird
}

func (gm *GirdMap) SetGirdEvent(key [2]int32, event events.Event) {
	oneGird := gm.girdMapData[key]
	oneGird.Events = append(oneGird.Events, event)
	gm.girdMapData[key] = oneGird
}

func (gm *GirdMap) GetGirdActors(key [2]int32) []int64 {
	val, ok := gm.girdMapData[key]
	if ok {
		return val.ActorIds
	}
	return nil
}

func (gm *GirdMap) GetGirdEvent(key [2]int32) []events.Event {
	val, ok := gm.girdMapData[key]
	if ok {
		return val.Events
	}
	return nil
}

func (gm *GirdMap) NineGrids(x, y int32) [][2]int32 {
	return [][2]int32{
		{x - 1, y - 1}, {x, y - 1}, {x + 1, y - 1},
		{x - 1, y}, {x, y}, {x + 1, y},
		{x - 1, y + 1}, {x, y + 1}, {x + 1, y + 1},
	}
}
