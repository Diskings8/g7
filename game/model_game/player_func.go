package model_game

import "fmt"

func (p *Player) GetSeverAndPlayerId() string {
	return fmt.Sprintf("%d_%d", p.ServerId, p.PlayerId)
}
