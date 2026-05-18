package tileds

import (
	"errors"
	"fmt"
	"g7/common/logger"
	"g7/common/maps/maps_data"
	"github.com/lafriks/go-tiled"
	"sync"
)

type TileTransform struct {
	mapD map[string]maps_data.MapData
	rw   sync.RWMutex
}

func (ttf *TileTransform) Init() {
	ttf.mapD = make(map[string]maps_data.MapData)
}

func (ttf *TileTransform) GetMap(name string) maps_data.MapData {
	ttf.rw.RLock()
	defer ttf.rw.RUnlock()
	return ttf.mapD[name]
}

func (ttf *TileTransform) ReadTiledMap(tileFile, mapName, wallName string) (err error) {
	//fmt.Println(filepath.Abs(tileFile))
	m, err := tiled.LoadFile(tileFile)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("%s load file error", tileFile))
		return err
	}
	var found bool
	var md = make([][2]int, 0)
	for _, l := range m.Layers {
		if l.Name != wallName {
			continue
		}
		found = true
		width := m.Width
		height := m.Height
		tiledSize := m.TileWidth
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				titleGirdId := l.Tiles[y*m.Width+x].ID
				if titleGirdId == 0 {
					continue
				}
				//fmt.Println(titleGirdId, y, x)
				md = append(md, [2]int{y, x})
			}
		}
		ttf.rw.Lock()
		ttf.mapD[mapName] = maps_data.NewMapData(mapName, width, height, tiledSize, md)
		ttf.rw.Unlock()
	}
	if !found {
		logger.Log.Warn(fmt.Sprintf("%s not found in tile file", wallName))
		return errors.New("not found title Wall")
	}
	return
}
