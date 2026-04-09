package env_conf

import (
	"fmt"
)

type MySQL struct {
	User         string `yaml:"user"`
	Pass         string `yaml:"pass"`
	Addr         string `yaml:"addr"`
	Port         string `yaml:"port"`
	DbNamePrefix string `yaml:"db_name_prefix"`
	Params       string `yaml:"params"`
}

func (ms *MySQL) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", ms.User, ms.Pass, ms.Addr, ms.Port, ms.DbNamePrefix, ms.Params)
}

func (ms *MySQL) DsnWithName(name string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s_%s?%s", ms.User, ms.Pass, ms.Addr, ms.Port, ms.DbNamePrefix, name, ms.Params)
}

type MongoDBConfig struct {
	URI          string `yaml:"uri"`
	DbNamePrefix string `yaml:"db_name_prefix"`
	PoolMin      int    `yaml:"pool_min"`
	PoolMax      int    `yaml:"pool_max"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Snowflake 雪花算法配置
type Snowflake struct {
	DatacenterID int64 `yaml:"datacenter_id"`
	WorkerID     int64 `yaml:"worker_id"`
}

type Server struct {
	Platform string `yaml:"platform"`
	Login    string `yaml:"login"`
	Game     string `yaml:"game"`
}

type JWT struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type Etcd struct {
	Dsn string `yaml:"dsn"`
}

type GateWay struct {
	Addr string `yaml:"addr"`
	Port string `yaml:"port"`
}

func (gw *GateWay) Dsn() string {
	return fmt.Sprintf("%s:%s", gw.Addr, gw.Port)
}

type Env struct {
	ResetHour       int `yaml:"reset_hour"`
	HeatBeatSeconds int `yaml:"heat_beat_seconds"`
}

type MQ struct {
	Dsn  string `yaml:"dsn"`
	Kind string `yaml:"kind"`
}

type Config struct {
	MySQLGlobal MySQL     `yaml:"mysql_global"`
	MySQLGame   MySQL     `yaml:"mysql_game"`
	Redis       Redis     `yaml:"redis"`
	Snowflake   Snowflake `yaml:"snowflake"`
	Server      Server    `yaml:"server"`
	JWT         JWT       `yaml:"jwt"`
	Etcd        Etcd      `yaml:"etcd"`
	GateWay     GateWay   `yaml:"gateWay"`
	Env         Env       `yaml:"env"`
	MQ          MQ        `yaml:"mq"`
}
