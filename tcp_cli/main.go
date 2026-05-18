package main

import (
	"flag"
	"fmt"
	"g7/common/maps/tileds"
)

func main() {
	flag.StringVar(&cmdParms, "role", "1", "")
	flag.Parse()

	//TC()
	tts := tileds.TileTransform{}
	tts.Init()
	fmt.Println(tts.ReadTiledMap("../../G7Conf/tiledmaps/maps/FlowerRoom.tmx", "FlowerRoom", "Wall"))
}
