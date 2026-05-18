package maps_data

import "math"

type MapData struct {
	MapName    string
	Width      int // 水平格子数
	Height     int // 垂直格子数
	MapWidth   int // 地图总像素宽
	MapHeight  int // 地图总像素高
	TileWidth  int // 单个格子的像素宽
	TileHeight int // 单个格子的像素高
	WallIds    map[int]struct{}
}

func NewMapData(name string, width, height int, tileSize int, wall [][2]int) MapData {
	md := MapData{MapName: name,
		MapWidth: width * tileSize, MapHeight: height * tileSize,
		Width: width, Height: height,
		TileHeight: tileSize,
		TileWidth:  tileSize,
		WallIds:    make(map[int]struct{})}
	for _, v := range wall {
		gx, gy := v[0], v[1]
		md.WallIds[gy*width+gx] = struct{}{}
	}
	return md
}

func (md MapData) IsBlockGrid(gx, gy int) bool {
	if gx < 0 || gy < 0 || gx >= md.Width || gy >= md.Height {
		return true
	}
	_, ok := md.WallIds[gy*md.Width+gx]
	return ok
}

// IsBlock 前端坐标
func IsBlock[T int | float32 | float64](md MapData, x, y T) bool {
	var gx, gy int
	switch any(x).(type) {
	case int:
		gx = int(x)
		gy = int(y)
	case float32, float64:
		gx = int(math.Floor(float64(x) / float64(md.TileWidth)))
		gridY := math.Floor(float64(y) / float64(md.TileHeight))
		gy = md.Height - 1 - int(gridY)
	}
	return md.IsBlockGrid(gx, gy)
}
