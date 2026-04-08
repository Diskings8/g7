package general_system_game

import (
	"encoding/json"
	"errors"
	"fmt"
	"g7/common/conf_data"
	"g7/common/confs"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/snowflakes"
	"g7/common/structs"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"go.uber.org/zap"
	"time"
)

var GBagSystem bagSystem

type bagSystem struct {
	bagTypeList []uint8
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_BagSystem, &bagSystem{})
}

func (this *bagSystem) Init() {
	this.bagTypeList = []uint8{const_game.BagType_Default, const_game.BagType_Currency}
}

func (this *bagSystem) GetName() string {
	return "general_bag_system"
}

func (this *bagSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {
	Player.AllBagData.Init()
	for _, bagType := range this.bagTypeList {
		Player.AllBagData.NewBag(bagType)
	}
	val := model_game.AllBagData{}
	_ = json.Unmarshal(dao.GeneralD.BagData, &val)
	for k, v := range val.Bags {
		Player.AllBagData.ReplaceBag(k, v)
	}
	// 检查过期道具
}

func (this *bagSystem) DailyReset(Player *model_game.Player) {}

func (this *bagSystem) OnEnterGame(Player *model_game.Player) {

}

func (this *bagSystem) GainAndConsumption(GainItemKV, CostItemKV []structs.KInt32VInt64, reason string, Player *model_game.Player) (bool, error) {

	costMap := this.splitByResourceType(CostItemKV)
	// 检查消耗
	for k, v := range costMap {
		bag := Player.AllBagData.GetBag(k)
		for _, vv := range v {
			if !bag.CheckCfgIdEnough(vv.K, vv.V) {
				return false, errors.New(fmt.Sprintf("%s not enough", conf_data.GetItemByID(vv.K).Name))
			}
		}
	}

	// 执行消耗
	for k, v := range costMap {
		bag := Player.AllBagData.GetBag(k)
		for _, vv := range v {
			bag.RemoveItemByCfgId(vv.K, vv.V)
		}
	}

	// 执行获得
	gainMap := this.splitByResourceType(GainItemKV)
	for k, v := range gainMap {
		bag := Player.AllBagData.GetBag(k)
		for _, vv := range v {
			confData := conf_data.GetItemByID(vv.K)
			item := this.newItem(confData, vv.V, Player.PlayerId)
			bag.AddItem(item)
		}
	}

	// 编写日志
	costCurrency, costOther := this.splitCurrencyBag(costMap)
	gainCurrency, gainOther := this.splitCurrencyBag(gainMap)
	actionLog := model_common.ActionLog{
		PlayerID:     Player.PlayerId,
		Action:       "GainAndConsumption",
		Reason:       reason,
		CostItem:     costOther,
		CostCurrency: costCurrency,
		GainItem:     gainOther,
		GainCurrency: gainCurrency,
		Ext:          "",
		CreateTime:   time.Now(),
	}
	Player.ActionLogs = append(Player.ActionLogs, &actionLog)

	return true, nil
}

func (this *bagSystem) newItem(cfg confs.Item, num int64, PlayerId int64) model_game.ItemData {
	val := model_game.ItemData{
		UniqueID:   snowflakes.GenUUID(),
		CfgID:      cfg.CfgID,
		IsBind:     cfg.IsBind,
		IsUnique:   cfg.IsUnique,
		ExpireType: cfg.ExpireType,
		Num:        num,
		OwnerID:    PlayerId,
		CreateID:   PlayerId,
		CreateTime: time.Time{},
	}
	if cfg.ExpireType == utils.TimeSpecified {
		val.ExpireTime = time.Unix(cfg.LimitTime, 0)
	} else if cfg.ExpireType == utils.TimeLimit {
		val.ExpireTime = time.Now().Add(time.Duration(cfg.LimitTime))
	}
	return val
}

// 根据配置表自动切割：货币 / 道具
func (this *bagSystem) splitByResourceType(rewards []structs.KInt32VInt64) map[uint8][]structs.KInt32VInt64 {
	bagMaps := make(map[uint8][]structs.KInt32VInt64)
	for _, rew := range rewards {
		cfg := conf_data.GetItemByID(rew.K)
		bagT, err := utils.Int32ToUint8(cfg.ResourceType)
		if err != nil {
			logger.Log.Warn("splitByResourceType fail", zap.Error(err))
			continue
		}
		bagMaps[bagT] = append(bagMaps[bagT], rew)
	}
	return bagMaps
}

func (this *bagSystem) splitCurrencyBag(src map[uint8][]structs.KInt32VInt64) ([]structs.KInt32VInt64, []structs.KInt32VInt64) {
	currencyBags := make([]structs.KInt32VInt64, 0)
	otherBags := make([]structs.KInt32VInt64, 0)
	for k, v := range src {
		if k == const_game.BagType_Currency {
			currencyBags = append(currencyBags, v...)
			continue
		}
		otherBags = append(otherBags, v...)
	}
	return currencyBags, otherBags
}
