package model_game

import "g7/common/logger"

type Player struct {
	OnlineData

	PlayerId int64
	Nickname string `json:"nickname"`
}

func (p *Player) Save() {
	logger.Log.Info("save player")
}
