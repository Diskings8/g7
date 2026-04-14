package mysql_driver

import (
	"fmt"
	"g7/common/dbc/dbc_interface"
	"g7/common/model_common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reflect"
)

type MySQLDriver struct {
	db *gorm.DB
	tx *MySQLTxDriver
}

func NewMySQLDriver(dsn string) (*MySQLDriver, error) {
	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	driver := &MySQLDriver{db: orm}
	return driver, nil
}

func (g *MySQLDriver) AutoMigrate(model model_common.DBTableInterface) error {
	return g.db.AutoMigrate(model)
}

func (m *MySQLDriver) Insert(model model_common.DBTableInterface) error {
	// collection = 表名
	// conf_data = 任意结构体
	return m.db.Table(model.TableName()).Save(model).Error
}

func (m *MySQLDriver) FindOne(table model_common.DBTableInterface, query any) error {
	return m.db.Table(table.TableName()).Where(query).First(table).Error
}

func (m *MySQLDriver) IsTableExists(tableName string) bool {
	var count int64
	err := m.db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_name = ?
	`, tableName).Scan(&count).Error
	return err == nil && count > 0
}

func (m *MySQLDriver) Exec(sql string) error {
	return m.db.Exec(sql).Error
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

func (m *MySQLTxDriver) BatchMQInsert(models []model_common.DBMqInterface) error {
	if len(models) == 0 {
		return nil
	}
	// 取第一个元素的真实类型
	first := models[0]
	val := reflect.ValueOf(first)
	slice := reflect.MakeSlice(reflect.SliceOf(val.Type()), len(models), len(models))
	for i := range models {
		slice.Index(i).Set(reflect.ValueOf(models[i]))
	}
	return m.tx.Table(first.TableName()).CreateInBatches(slice.Interface(), len(models)).Error
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
