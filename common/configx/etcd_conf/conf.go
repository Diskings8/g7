package etcd_conf

import "encoding/json"

type Config struct {
	ServerEnv      string  // 环境 test/prod
	LogLevel       string  // 日志级别
	RegisterOn     bool    // 注册开关
	LoginOn        bool    // 登录开关
	RechargeOn     bool    // 充值开关
	CrossOn        bool    // 跨服开关
	ExpRate        float64 // 经验倍率
	DropRate       float64 // 掉落倍率
	ProtocolEncode bool    // 协议加密
	LimitApiFreq   int     // 接口限流
}

const (
	ConfGlobalServerEnv  = "/config/global/server_env"
	ConfGlobalLogLevel   = "/config/global/log_level"
	ConfSwitchRegisterOn = "/config/switch/register_enable"
	ConfSwitchLoginOn    = "/config/switch/login_enable"
	ConfSwitchRechargeOn = "/config/switch/recharge_enable"
	ConfSwitchCrossOn    = "/config/switch/cross_enable"
	ConfSwitchExpRate    = "/config/hotfix/exp_rate"
	ConfSwitchDropRate   = "/config/hotfix/drop_rate"
	ConfSwitchProtocolOn = "/config/switch/protocol_encode"
)

func (conf *Config) SetConf(key, value string) {
	switch key {
	case ConfGlobalServerEnv:
		conf.ServerEnv = value
	case ConfGlobalLogLevel:
		conf.LogLevel = value
	case ConfSwitchRegisterOn:
		conf.RegisterOn = value == "true"
	case ConfSwitchLoginOn:
		conf.LoginOn = value == "true"
	case ConfSwitchRechargeOn:
		conf.RechargeOn = value == "true"
	case ConfSwitchCrossOn:
		conf.CrossOn = value == "true"
	case ConfSwitchExpRate:
		_ = json.Unmarshal([]byte(value), &conf.ExpRate)
	case ConfSwitchDropRate:
		_ = json.Unmarshal([]byte(value), &conf.DropRate)
	case ConfSwitchProtocolOn:
		conf.ProtocolEncode = value == "true"
	}
}
