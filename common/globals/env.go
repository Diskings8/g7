package globals

// Env 全局环境变量
var (
	Env        string // 开发环境
	ServerId   string // 服务器id
	InstanceId string // 实例
	Container  string // 容器
	Platform   string //
)

const (
	ConfEnv  = "../config.yaml"
	ConfProd = "./config.yaml"
	ConfPre  = "../config_prod.yaml"
)

const (
	EnvTest = "test"
	EnvProd = "prod"
	EnvPre  = "pre"

	ContainerLocal  = "local"
	ContainerDocker = "docker"
)

func GetEnvConfPath() string {
	switch Env {
	case EnvProd:
		return ConfProd
	case EnvPre:
		return ConfPre
	default:
		return ConfEnv
	}
}

func IsDev() bool {
	return Env == EnvTest
}

func IsProd() bool {
	return Env == EnvProd
}

func IsContainerLocal() bool {
	return Container == ContainerLocal
}

func IsContainerDocker() bool {
	return Container == ContainerDocker
}
