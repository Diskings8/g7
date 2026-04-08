package conf_data

import "g7/common/confs"

func GetItemByID(id int32) confs.Item {
	switch id {
	case 10001:
		return confs.Item{
			CfgID:           id,
			ResourceType:    1,
			ResourceSubType: 1,
			Name:            "铜钱",
		}
	case 10002:
		return confs.Item{
			CfgID:           id,
			ResourceType:    1,
			ResourceSubType: 2,
			Name:            "砖石",
		}
	case 20001:
		return confs.Item{
			CfgID:        id,
			ResourceType: 2,
			Name:         "经验药水",
		}
	case 20002:
		return confs.Item{
			CfgID:        id,
			ResourceType: 2,
			Name:         "回复药水",
		}
	case 30001:
		return confs.Item{
			CfgID:        id,
			ResourceType: 2,
			Name:         "充值礼包",
		}
	}
	return confs.Item{
		CfgID:           id,
		ResourceType:    2,
		ResourceSubType: 1,
		Name:            "未配置",
	}
}
