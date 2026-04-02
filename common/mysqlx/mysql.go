package mysqlx

import (
	"g7/common/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var GlobalDb *gorm.DB
var GameServerDb *gorm.DB

func InitGlobalDb(dsn string) {
	GlobalDb = initMySQL(dsn)
}

func InitGameServerDb(dsn string) {
	GameServerDb = initMySQL(dsn)
}

func initMySQL(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("登录服MySQL连接失败", zap.Error(err))
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Log.Fatal("获取DB连接池失败", zap.Error(err))
		panic(err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

func AutoMigrate(db *gorm.DB, dst ...interface{}) {
	err := db.AutoMigrate(dst...)
	if err != nil {
		logger.Log.Fatal("自动建表失败", zap.Error(err))
		panic(err)
	}
}
