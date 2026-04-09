package dbc

import (
	"g7/common/dbc/dbc_interface"
	"g7/common/dbc/mongo_driver"
	"g7/common/dbc/mysql_driver"
	"g7/common/logger"
	"g7/common/model_common"
	"log"
)

// InitDB 启动时根据配置初始化
func InitDB(dbType string, dsn string) dbc_interface.DBInterface {
	switch dbType {
	case "mongo":
		mongo, err := mongo_driver.NewMongoDriver(dsn)
		if err != nil {
			panic(err)
		}
		log.Println("使用 MongoDB 存储")
		return mongo
	case "mysql":
		mysql, err := mysql_driver.NewMySQLDriver(dsn)
		if err != nil {
			panic(err)
		}
		log.Println("使用 MySQL 存储")
		return mysql
	default:
		panic("不支持的数据库类型")
	}
}

func AutoMigrates(db dbc_interface.DBInterface, models ...model_common.DBTableInterface) error {
	for _, m := range models {
		err := db.AutoMigrate(m)
		if err != nil {
			logger.Log.Warn(m.TableName() + "AutoMigrate failed!")
			continue
		}
	}
	return nil
}
