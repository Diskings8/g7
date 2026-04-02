package dao_login

import (
	"g7/common/mysqlx"
	"g7/common/utils"
	"g7/login/model_login"
)

func AutoMigrate() {
	if utils.IsDev() {
		mysqlx.AutoMigrate(mysqlx.GlobalDb, &model_login.User{}, &model_login.Player{})
	}
}
