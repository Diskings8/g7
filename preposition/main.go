package main

import (
	"g7/common/configx"
	"g7/common/configx/env_conf"
	"g7/common/dbc"
	"g7/common/dbc/dbc_interface"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
)

func main() {

	confStr := globals.ConfDev
	env_conf.Load(confStr)

	logger.Init()

	var dbT dbc_interface.DBInterface

	dbT = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.Dsn())
	_ = dbc.AutoMigrates(dbT, &model_common.Server{}, &model_common.GlobalPlayerIndex{})

}
