package manager_game

import (
	"fmt"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/game/global_game"
	"g7/game/interface_game"
	"sync"
	"sync/atomic"
	"time"
)

var GActivityManager = &activityManager{}

type activityManager struct {
	activityData    map[int32][]*model_common.GameActivity
	activitySystems map[int32]interface_game.ActivitySystem
	isChecking      int32
	rwLock          sync.RWMutex
	timeWheel       *timeWheel // 统一时间轮
}

type timeWheel struct {
	ticker *time.Ticker
}

func (m *activityManager) Init() {
	m.activityData = make(map[int32][]*model_common.GameActivity)
	m.activitySystems = make(map[int32]interface_game.ActivitySystem)

}

func (m *activityManager) Register(systemId int32, sys interface_game.ActivitySystem) {
	m.rwLock.RLock()
	_, ok := m.activitySystems[systemId]
	m.rwLock.RUnlock()

	if ok {
		logger.Log.Warn(fmt.Sprintf("%d system had been Register in ActivitySystem", systemId))
		return
	}

	m.rwLock.Lock()
	m.activitySystems[systemId] = sys
	m.rwLock.Unlock()

	sys.Init()
}

func (m *activityManager) InitData() {
	var allActivity []*model_common.GameActivity
	err := global_game.GGameDB.FindList(&allActivity, "delete_flag != ?", []string{"1"})
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	// 先同步本服的活动
	m.rwLock.Lock()
	for _, v := range allActivity {
		m.activityData[v.ConfId] = append(m.activityData[v.ConfId], v)
	}
	m.rwLock.Unlock()
	m.activityStateCheck()
	go m.timerActivityStateCheck()
}

func (m *activityManager) activityStateCheck() {
	if !atomic.CompareAndSwapInt32(&m.isChecking, 0, 1) {
		// 已经在执行了，直接退出，不重复执行
		return
	}
	defer atomic.StoreInt32(&m.isChecking, 0)

	now := time.Now().Unix()
	var needOpenActivity, needFinishActivity, needCloseActivity []*model_common.GameActivity

	m.rwLock.RLock()
	for _, acts := range m.activityData {
		for _, act := range acts {
			if act.ActivityType == globals.ActivityTypePermanent && act.Status != globals.ActivityStateOpen {
				needOpenActivity = append(needOpenActivity, act)
				continue
			}
			if act.Status == globals.ActivityStateClose && act.StartTime <= now && now < act.EndTime {
				needOpenActivity = append(needOpenActivity, act)
				continue
			}
			if act.Status == globals.ActivityStateOpen && act.EndTime <= now {
				needFinishActivity = append(needFinishActivity, act)
				continue
			}
			if (act.Status == globals.ActivityStateOpen || act.Status == globals.ActivityStateFinish) && act.CloseTime <= now {
				needCloseActivity = append(needCloseActivity, act)
				continue
			}
		}
	}
	m.rwLock.RUnlock()
	{
		m.batchUpdateMemoryStatus(needOpenActivity, globals.ActivityStateOpen)
		m.batchUpdateMemoryStatus(needFinishActivity, globals.ActivityStateFinish)
		m.batchUpdateMemoryStatus(needCloseActivity, globals.ActivityStateClose)
	}
	{
		m.batchUpdateDBActivityStatus(needOpenActivity, globals.ActivityStateOpen)
		m.batchUpdateDBActivityStatus(needFinishActivity, globals.ActivityStateFinish)
		m.batchUpdateDBActivityStatus(needCloseActivity, globals.ActivityStateClose)
	}

	m.triggerActivityEvents(needOpenActivity, needFinishActivity, needCloseActivity)
}

func (m *activityManager) batchUpdateMemoryStatus(acts []*model_common.GameActivity, status int32) {
	if len(acts) <= 0 {
		return
	}
	for _, act := range acts {
		act.Status = status
	}
}

// 批量更新活动状态到 DB（推荐！）
func (m *activityManager) batchUpdateDBActivityStatus(acts []*model_common.GameActivity, status int32) {
	if len(acts) == 0 {
		return
	}

	// 收集所有活动 ID
	var ids []int64
	for _, act := range acts {
		ids = append(ids, act.ActivityId)
	}

	// 🔥 批量 UPDATE 语句（只执行一次）
	err := global_game.GGameDB.Update(&model_common.GameActivity{}, map[string]any{"status": status}, "id IN (?)", ids)
	if err != nil {
		logger.Log.Error("batch update activity status fail: " + err.Error())
	}
}

// 触发活动开启/结束/关闭事件
func (m *activityManager) triggerActivityEvents(needOpenActivity, needFinishActivity, needCloseActivity []*model_common.GameActivity) {
	if len(needOpenActivity) > 0 {
		for _, act := range needOpenActivity {
			m.rwLock.RLock()
			sys, ok := m.activitySystems[act.ConfId]
			m.rwLock.RUnlock()
			if !ok {
				continue
			}
			go func(confId int32, actId int64) {
				sys.OnActivityOpen(confId, actId)
			}(act.ConfId, act.ActivityId)
		}
	}
	if len(needFinishActivity) > 0 {
		for _, act := range needFinishActivity {
			m.rwLock.RLock()
			sys, ok := m.activitySystems[act.ConfId]
			m.rwLock.RUnlock()
			if !ok {
				continue
			}
			go func(confId int32, actId int64) {
				sys.OnActivityFinish(confId, actId)
			}(act.ConfId, act.ActivityId)
		}
	}
	if len(needCloseActivity) > 0 {
		for _, act := range needCloseActivity {
			m.rwLock.RLock()
			sys, ok := m.activitySystems[act.ConfId]
			m.rwLock.RUnlock()
			if !ok {
				continue
			}
			go func(confId int32, actId int64) {
				sys.OnActivityClose(confId, actId)
			}(act.ConfId, act.ActivityId)
		}
	}
}

func (m *activityManager) timerActivityStateCheck() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop() // 加上这一行，优雅停止，防止内存泄漏
	for range ticker.C {
		m.activityStateCheck()
	}
}
