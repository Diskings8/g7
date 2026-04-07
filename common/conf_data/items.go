package conf_data

import "g7/common/confs"

func GetItemByID(id int32) confs.Item {
	return confs.Item{
		CfgID: id,
	}
}
