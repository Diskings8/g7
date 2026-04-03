package main

import (
	"g7/common/config"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/mysqlx"
)

func main() {

	confStr := globals.ConfDev
	config.Load(confStr)

	logger.Init()

	mysqlx.InitGlobalDb(config.GCfg.MySQLGlobal.Dsn())
	mysqlx.AutoMigrate(mysqlx.GlobalDb, &model_common.Server{}, &model_common.GlobalPlayerIndex{})

}
