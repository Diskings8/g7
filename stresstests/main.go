package main

import (
	"fmt"
	"g7/stresstests/autobot"
	"g7/stresstests/createPlayer"
	"sync"
)

func main() {
	//autobot.ConnectToDb()
	autobot.LoadAndConn()
}

func countCreatePlayer() {
	wg := sync.WaitGroup{}
	var tCount = int(1000)
	wg.Add(tCount)
	for i := 1; i <= tCount; i++ {
		go func() {
			defer wg.Done()
			createPlayer.OneConnect(i)
		}()
	}
	wg.Wait()
	fmt.Println(createPlayer.GSuccess, createPlayer.GFail)
}
