package dao_login

import (
	"g7/common/model_common"
	"g7/login/global_login"
)

func GetServerByID(serverID int32) (*model_common.Server, error) {
	var s model_common.Server
	err := global_login.GLoginDB.FindOne(&s, map[string]interface{}{"server_id": serverID})
	return &s, err
}

func ListServersByChannel(channel int32) ([]*model_common.Server, error) {
	var list []*model_common.Server
	err := global_login.GLoginDB.FindList(&list, map[string]interface{}{"channel": channel})
	return list, err
}
