package mysqlx

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

func Init(dsn string) {
	db, err := gorm.Open(mysql.Open(dsn), nil)
	if err != nil {
		panic("mysql 连接失败: " + err.Error())
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
}
