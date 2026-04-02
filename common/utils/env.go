package utils

import (
	"g7/common/globals"
)

const (
	EnvTest = "test"
)

func IsDev() bool {
	return globals.Env == EnvTest
}
