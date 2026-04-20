package global_login

import (
	"g7/common/dbc"
	"g7/common/dbc/dbc_interface"
	"g7/common/model_common"
	"g7/login/model_login"
)

var GLoginDB dbc_interface.DBInterface

func AutoMigrate(dbi dbc_interface.DBInterface) {
	_ = dbc.AutoMigrates(dbi, &model_login.User{}, &model_common.GameOrder{},
		&model_common.PaymentRecord{}, &model_common.BaseMail{},
		&model_common.BaseActivity{})
}
