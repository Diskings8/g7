package mysql_driver

import (
	"g7/common/model_common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLDriver struct {
	db *gorm.DB
}

func NewMySQLDriver(dsn string) (*MySQLDriver, error) {
	orm, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	driver := &MySQLDriver{db: orm}
	return driver, nil
}

func (g *MySQLDriver) AutoMigrate(model model_common.DBTableInterface) error {
	return g.db.Table(model.TableName()).AutoMigrate(model)
}

func (m *MySQLDriver) Insert(model model_common.DBTableInterface) error {
	// collection = 表名
	// conf_data = 任意结构体
	return m.db.Table(model.TableName()).Save(model).Error
}

func (m *MySQLDriver) FindOne(table model_common.DBTableInterface, query any) error {
	return m.db.Table(table.TableName()).Where(query).First(table).Error
}

// --------------------
// FindList 查询列表
// --------------------
func (m *MySQLDriver) FindList(result any, query any) error {
	return m.db.Where(query).Find(result).Error
}
