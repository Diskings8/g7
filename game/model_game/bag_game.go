package model_game

import (
	"g7/common/utils"
)

type AllBagData struct {
	Bags map[uint8]*Bag `json:"bags"`
}

func (ab *AllBagData) Init() {
	ab.Bags = make(map[uint8]*Bag)
}

func (ab *AllBagData) NewBag(bagType uint8) {
	ab.Bags[bagType] = new(Bag)
	ab.Bags[bagType].Init()
}

func (ab *AllBagData) GetBag(bagType uint8) *Bag {
	return ab.Bags[bagType]
}

func (ab *AllBagData) ReplaceBag(bagType uint8, bag *Bag) {
	ab.Bags[bagType] = bag
}

type Bag struct {
	BagItems map[uint64]*ItemData `json:"bag_items"` // uuid
	BagSize  int                  `json:"bag_size"`
}

func (this *Bag) Init() {
	this.BagItems = make(map[uint64]*ItemData)
}

func (b *Bag) FindOneByCfgID(CfgId int32) *ItemData {
	for _, item := range b.BagItems {
		if item.CfgID == CfgId {
			return item
		}
	}
	return nil
}

func (b *Bag) FindAllByCfgID(CfgId int32) []*ItemData {
	val := make([]*ItemData, 0)
	for _, item := range b.BagItems {
		if item.CfgID == CfgId {
			val = append(val, item)
		}
	}
	return val
}

func (b *Bag) CheckCfgIdEnough(CfgId int32, num int64) bool {
	allCheckNum := int64(0)
	for _, item := range b.BagItems {
		if item.CfgID == CfgId {
			allCheckNum += item.Num
		}
	}
	return allCheckNum >= num
}

func (b *Bag) FindOneByUniqueID(UUID uint64) *ItemData {
	return b.BagItems[UUID]
}

func (b *Bag) AddItem(newItem ItemData) {
	for k, item := range b.BagItems {
		if item.CfgID == newItem.CfgID &&
			item.IsUnique == newItem.IsUnique &&
			item.IsBind == newItem.IsBind &&
			item.ExpireType == newItem.ExpireType {
			item.Num += newItem.Num
			b.BagItems[k] = item
			return
		}
	}
	b.BagItems[newItem.UniqueID] = &newItem
	return
}

func (b *Bag) RemoveItemByCfgId(cfgId int32, num int64) bool {
	canUseBindKey := make([]uint64, 0)
	canUseNoBindKey := make([]uint64, 0)
	var curNum int64
	for k, item := range b.BagItems {
		if item.CfgID == cfgId {
			curNum += item.Num
			if item.IsBind == utils.ConstOne {
				canUseBindKey = append(canUseBindKey, k)
			} else {
				canUseNoBindKey = append(canUseNoBindKey, k)
			}
		}
	}
	if curNum < num {
		return false
	}
	reqNum := num
	if len(canUseBindKey) > 0 {
		for _, v := range canUseBindKey {
			if b.BagItems[v].Num >= reqNum {
				b.BagItems[v].Num -= reqNum
				return true
			} else {
				reqNum -= b.BagItems[v].Num
				delete(b.BagItems, v)
			}
		}
	}
	if len(canUseNoBindKey) > 0 {
		for _, v := range canUseNoBindKey {
			if b.BagItems[v].Num >= reqNum {
				b.BagItems[v].Num -= reqNum
				return true
			} else {
				reqNum -= b.BagItems[v].Num
				delete(b.BagItems, v)
			}
		}
	}
	return true
}

func (b *Bag) CheckItemEnough(cfgId int32, num int64) bool {
	var curNum int64
	for _, item := range b.BagItems {
		if item.CfgID == cfgId {
			curNum += item.Num
		}
	}
	return curNum >= num
}
