package dbc_interface

import "g7/common/model_common"

// 玩家数据结构（两套数据库共用）

// DBInterface 上层业务只认这个接口
type DBInterface interface {
	Insert(model model_common.DBTableInterface) error
	AutoMigrate(model model_common.DBTableInterface) error
	FindOne(table model_common.DBTableInterface, query any) error // 查询单条，结果存入 table
	FindList(table any, query any) error                          // 查询列表，结果存入 table
	IsTableExists(tableName string) bool
	Begin() DBTxInterface // 或者用 interface{} 做泛型，这里用具体类型更简单

}

type DBTxInterface interface {
	Commit() error
	Rollback() error
	BatchMQInsert(model []model_common.DBMqInterface) error
}
