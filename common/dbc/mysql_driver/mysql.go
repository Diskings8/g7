package mysql_driver

import (
	"fmt"
	"g7/common/dbc/dbc_interface"
	"g7/common/model_common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reflect"
	"time"
)

type MySQLDriver struct {
	db *gorm.DB
	tx *gorm.DB
}

func (m *MySQLDriver) getDb() *gorm.DB {
	//if globals.IsDev() {
	//	return m.db.Debug()
	//}
	if m.tx != nil {
		return m.tx
	}
	return m.db
}

func NewMySQLDriver(dsn string) (*MySQLDriver, error) {
	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := orm.DB()
	if err != nil {
		return nil, err
	}

	// 配置连接池（根据你的服务器配置调整）
	sqlDB.SetMaxOpenConns(100)                 // 最大同时打开的连接数
	sqlDB.SetMaxIdleConns(50)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(30 * time.Second) // 连接最长存活时间
	sqlDB.SetConnMaxIdleTime(10 * time.Second) // 空闲连接最长存活时间

	driver := &MySQLDriver{db: orm}
	return driver, nil
}

func (m *MySQLDriver) AutoMigrate(model model_common.DBTableInterface) error {
	return m.db.AutoMigrate(model)
}

func (m *MySQLDriver) Delete(model model_common.DBTableInterface) error {

	return m.getDb().Table(model.TableName()).Delete(model).Error
}

func (m *MySQLDriver) Insert(model model_common.DBTableInterface) error {
	// collection = 表名
	// conf_data = 任意结构体
	return m.getDb().Table(model.TableName()).Save(model).Error
}

func (m *MySQLDriver) BatchInsert(models []model_common.DBTableInterface) error {
	first := models[0]
	val := reflect.ValueOf(first)
	slice := reflect.MakeSlice(reflect.SliceOf(val.Type()), len(models), len(models))
	for i := range models {
		slice.Index(i).Set(reflect.ValueOf(models[i]))
	}
	return m.getDb().Table(first.TableName()).CreateInBatches(slice.Interface(), len(models)).Error
}

func (m *MySQLDriver) Update(model model_common.DBTableInterface, updates any, query any, args ...any) error {
	return m.getDb().Model(model.TableName()).Where(query, args...).Updates(updates).Error
}

func (m *MySQLDriver) FindOne(table model_common.DBTableInterface, query any) error {
	return m.getDb().Table(table.TableName()).Where(query).First(table).Error
}

func (m *MySQLDriver) IsTableExists(tableName string) bool {
	var count int64
	err := m.getDb().Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_name = ?
	`, tableName).Scan(&count).Error
	return err == nil && count > 0
}

func (m *MySQLDriver) Exec(sql string) error {
	return m.getDb().Exec(sql).Error
}

func (m *MySQLDriver) FindList(result any, query any, params ...any) error {
	return m.getDb().Where(query, params...).Find(result).Error
}

func (m *MySQLDriver) FindListPro(result any, query any, order string, size, page int) error {
	return m.getDb().Where(query).Order(order).Limit(size).Offset((page - 1) * size).Find(result).Error
}

func (m *MySQLDriver) TxBegin() dbc_interface.DBInterface {
	return &MySQLDriver{tx: m.db.Begin()}
}

func (m *MySQLDriver) TxBatchMQInsert(models []model_common.DBMqInterface) error {
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

func (m *MySQLDriver) TxCommit() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction started")
	}

	return m.tx.Commit().Error
}

func (m *MySQLDriver) TxRollback() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction started")
	}
	return m.tx.Rollback().Error
}
