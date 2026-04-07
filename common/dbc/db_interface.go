package dbc

import "g7/common/model_common"

// 玩家数据结构（两套数据库共用）

// DBInterface 上层业务只认这个接口
type DBInterface interface {
	Insert(model model_common.DBTableInterface) error
	AutoMigrate(model model_common.DBTableInterface) error
	FindOne(table model_common.DBTableInterface, query any) error // 查询单条，结果存入 table
	FindList(table any, query any) error                          // 查询列表，结果存入 table
}
