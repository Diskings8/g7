package service_login

import (
	"g7/common/model_common"
	"g7/login/internal/dao_login"
)

func GetServerByID(serverID int32) (*model_common.Server, error) {
	return dao_login.GetServerByID(serverID)
}
