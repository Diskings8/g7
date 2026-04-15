package main

import (
	"flag"
	"fmt"
	"g7/common/configx"
	"g7/common/configx/env_conf"
	"g7/common/configx/etcd_conf"
	"g7/common/dbc"
	"g7/common/dbc/dbc_interface"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
)

func main() {

	flag.StringVar(&globals.Env, "env", "test", "运行环境: test/prod/pre")
	flag.StringVar(&globals.InstanceId, "instance", "0", "实例id")
	flag.StringVar(&globals.Container, "container", "local", "容器类型：local/docker")
	flag.Parse()

	var confStr = globals.GetEnvConfPath()
	configx.LoadEnvConf(confStr)

	logger.Init()

	doAutoCreateDB()
	doAutoMigratesTable()
	doAutoSetEtcdConf()
}

func doAutoSetEtcdConf() {
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	etcd.UpdateEtcdConf(etcd_conf.ConfSwitchRegisterOn, "true")
	etcd.UpdateEtcdConf(etcd_conf.ConfSwitchLoginOn, "true")
}

func doAutoMigratesTable() {
	var dbT dbc_interface.DBInterface
	dbT = dbc.InitDB(globals.DBMysql, configx.GEnvCfg.MySQLGlobal.Dsn())
	_ = dbc.AutoMigrates(dbT, &model_common.Server{}, &model_common.GlobalPlayerIndex{})
}

func doAutoCreateDB() {
	p := configx.GEnvCfg.MySQLGlobal
	pp := env_conf.MySQL{
		User:         p.User,
		Pass:         p.Pass,
		Addr:         p.Addr,
		Port:         p.Port,
		DbNamePrefix: "mysql",
		Params:       p.Params,
	}
	var dbT dbc_interface.DBInterface
	dbT = dbc.InitDB(globals.DBMysql, pp.Dsn())

	dbName := configx.GEnvCfg.MySQLGlobal.DbNamePrefix

	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName)
	err := dbT.Exec(createDBSQL)
	if err != nil {
		panic("创建数据库失败: " + err.Error())
	}
	println("数据库创建成功！")
}
