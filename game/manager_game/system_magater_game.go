package manager_game

import (
	"fmt"
	"g7/common/logger"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/interface_game"
	"g7/game/model_game"
	"time"
)

var GISystemManager = &iSystemManager{
	ISystems: make(map[int32]interface_game.ISystem),
}
var GSaveSystemManager = &saveSystemManager{
	SaveSystems:    make(map[int32]interface_game.SaveSystem, 0),
	AsyncSaveQueue: make(chan *model_game.PlayerDao, 10000),
}

type iSystemManager struct {
	ISystems map[int32]interface_game.ISystem // 所有注册的系统
}

func (m *iSystemManager) Init() {

}

func (m *iSystemManager) Register(systemId int32, sys interface_game.ISystem) {
	if _, ok := m.ISystems[systemId]; ok {
		//logger.Log.Warn(fmt.Sprintf("%d system had been Register", systemId))
		return
	}
	m.ISystems[systemId] = sys
}

func (m *iSystemManager) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {
	// 优先加载玩家基础数据
	m.ISystems[const_game.General_PlayerSystem].LoadData(dao, Player)

	for id, sys := range m.ISystems {
		if id == const_game.General_PlayerSystem {
			continue
		}
		sys.LoadData(dao, Player)
	}
	return
}

func (m *iSystemManager) DailyReset(Player *model_game.Player) {
	// 判断是否是同一天
	if utils.CheckTwoTimeIsSameDay(Player.LastOfflineAt, time.Now()) {
		return
	}
	if sys, ok := m.ISystems[const_game.General_PlayerSystem].(interface_game.ISystem); ok {
		sys.DailyReset(Player)
	}
	for id, sys := range m.ISystems {
		if id == const_game.General_PlayerSystem {
			continue
		}
		sys.DailyReset(Player)
	}
	return
}

func (m *iSystemManager) OnEnterGame(Player *model_game.Player) {
	m.ISystems[const_game.General_PlayerSystem].OnEnterGame(Player)

	for id, sys := range m.ISystems {
		if id == const_game.General_PlayerSystem {
			continue
		}
		sys.OnEnterGame(Player)
	}
	return
}

type saveSystemManager struct {
	SaveSystems    map[int32]interface_game.SaveSystem // 所有注册的系统
	AsyncSaveQueue chan *model_game.PlayerDao
}

func (m *saveSystemManager) Init() {
	go func() {
		for dao := range m.AsyncSaveQueue {
			// 单协程写库
			m.SavePlayerDao(dao)
		}
	}()
}

func (m *saveSystemManager) Register(systemId int32, sys interface_game.SaveSystem) {
	if _, ok := m.SaveSystems[systemId]; ok {
		logger.Log.Warn(fmt.Sprintf("%d system had been Register", systemId))
		return
	}
	m.SaveSystems[systemId] = sys
}

func (m *saveSystemManager) SavePlayerDao(dao *model_game.PlayerDao) {
	for _, sys := range m.SaveSystems {
		sys.SavePlayerDao(dao)
	}
	return
}
