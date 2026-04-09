package mysql_driver

import (
	"fmt"
	"g7/common/dbc/dbc_interface"
	"g7/common/model_common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLDriver struct {
	db *gorm.DB
	tx *MySQLTxDriver
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

func (m *MySQLDriver) Begin() dbc_interface.DBTxInterface {
	m.tx = &MySQLTxDriver{tx: m.db.Begin()}
	return m.tx
}

type MySQLTxDriver struct {
	tx *gorm.DB
}

func (m *MySQLTxDriver) BatchMQInsert(model []model_common.DBMqInterface) error {
	if len(model) == 0 {
		return nil
	}
	tableName := model[0].TableName()
	return m.tx.Table(tableName).CreateInBatches(model, len(model)).Error
}

func (m *MySQLTxDriver) Commit() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction started")
	}

	return m.tx.Commit().Error
}

func (m *MySQLTxDriver) Rollback() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction started")
	}
	return m.tx.Rollback().Error
}
