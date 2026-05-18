package maps

import "g7/common/maps/tileds"

var GTiledManager tileds.TileTransform

func LoadAllMaps() {
	GTiledManager.Init()
	GTiledManager.ReadTiledMap("../../G7Conf/tiledmaps/maps/FlowerRoom.tmx", "FlowerRoom", "Wall")
}
