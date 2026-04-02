package dao_login

import (
	"g7/common/model_common"
	"g7/common/mysqlx"
)

func GetServerByID(serverID int) (*model_common.Server, error) {
	var s model_common.Server
	err := mysqlx.GlobalDb.Where("server_id = ? AND status = 1", serverID).First(&s).Error
	return &s, err
}

func ListServersByChannel(channel int) ([]*model_common.Server, error) {
	var list []*model_common.Server
	err := mysqlx.GlobalDb.Where("channel = ? OR channel = 0", channel).Find(&list).Error
	return list, err
}
