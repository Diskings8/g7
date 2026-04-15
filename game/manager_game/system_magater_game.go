package manager_game

import (
	"fmt"
	"g7/common/cronx"
	"g7/common/logger"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/global_game"
	"g7/game/interface_game"
	"g7/game/model_game"
	"time"
)

var GISystemManager = &iSystemManager{
	ISystems: make(map[int32]interface_game.ISystem),
}

var GSaveSystemManager = &saveSystemManager{
	SaveSystems:    make(map[int32]interface_game.SaveSystem, 0),
	AsyncSaveQueue: make(chan *model_game.SaveDaoD, 10000),
}

var GResetSystemManager = &resetSystemManager{
	ResetSystems: make(map[int32]interface_game.ResetSystem),
}

type iSystemManager struct {
	ISystems map[int32]interface_game.ISystem // 所有注册的系统
}

func (m *iSystemManager) Init() {

}

func (m *iSystemManager) Register(systemId int32, sys interface_game.ISystem) {
	if _, ok := m.ISystems[systemId]; ok {
		logger.Log.Warn(fmt.Sprintf("%d system had been Register in iSystemManager", systemId))
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
	AsyncSaveQueue chan *model_game.SaveDaoD
}

func (m *saveSystemManager) Init() {
	go func() {
		for dao := range m.AsyncSaveQueue {
			// 单协程写库
			m.SavePlayerDaoD(dao)
		}
	}()
}

func (m *saveSystemManager) Register(systemId int32, sys interface_game.SaveSystem) {
	if _, ok := m.SaveSystems[systemId]; ok {
		logger.Log.Warn(fmt.Sprintf("%d system had been Register in saveSystemManager", systemId))
		return
	}
	m.SaveSystems[systemId] = sys
}

func (m *saveSystemManager) SavePlayerDaoD(dao *model_game.SaveDaoD) {
	for _, sys := range m.SaveSystems {
		sys.SavePlayerDao(dao)
	}
	return
}

type resetSystemManager struct {
	ResetSystems map[int32]interface_game.ResetSystem // 所有注册的系统
}

func (m *resetSystemManager) Init() {
	cronx.AddDaily5HourTask(func() {
		global_game.GPlayerMaps.AllRunFunc(func(p *model_game.Player) {
			m.AllReset(p)
		})
	})
}

func (m *resetSystemManager) Register(systemId int32, sys interface_game.ResetSystem) {
	if _, ok := m.ResetSystems[systemId]; ok {
		logger.Log.Warn(fmt.Sprintf("%d system had been Register in resetSystemManager", systemId))
		return
	}
	m.ResetSystems[systemId] = sys
}

func (m *resetSystemManager) AllReset(Player *model_game.Player) {
	// 判断是否是同一天
	//logger.Log.Warn(fmt.Sprintf("%s,%s,%v", Player.LastDailyResetAt, time.Now(), utils.CheckTwoTimeIsSameDay(*Player.LastDailyResetAt, time.Now())))
	if !utils.CheckTwoTimeIsSameDay(Player.LastDailyResetAt, time.Now()) {
		// 先让其他系统处理每日刷新数据
		for id, sys := range m.ResetSystems {
			if id == const_game.General_PlayerSystem {
				continue
			}
			sys.DailyReset(Player)
		}
		// 更新玩家系统的每日数据
		if sys, ok := m.ResetSystems[const_game.General_PlayerSystem].(interface_game.ResetSystem); ok {
			sys.DailyReset(Player)
		}
	}
	// 判断是否同一周
	if !utils.CheckTwoTimeIsSameWeek(Player.LastWeekResetAt, time.Now()) {
		// 先让其他系统处理每日刷新数据
		for id, sys := range m.ResetSystems {
			if id == const_game.General_PlayerSystem {
				continue
			}
			sys.WeekReset(Player)
		}
		// 更新玩家系统的每日数据
		if sys, ok := m.ResetSystems[const_game.General_PlayerSystem].(interface_game.ResetSystem); ok {
			sys.WeekReset(Player)
		}
	}

	// 判断是否同一月
	if !utils.CheckTwoTimeIsSameMonth(Player.LastMonthResetAt, time.Now()) {
		// 先让其他系统处理每日刷新数据
		for id, sys := range m.ResetSystems {
			if id == const_game.General_PlayerSystem {
				continue
			}
			sys.MonthReset(Player)
		}
		// 更新玩家系统的每日数据
		if sys, ok := m.ResetSystems[const_game.General_PlayerSystem].(interface_game.ResetSystem); ok {
			sys.MonthReset(Player)
		}
	}

	return
}
