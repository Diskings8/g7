package utils

import "time"

const (
	ConstZero = iota
	ConstOne
)

const (
	TimeNoLimit = iota
	TimeSpecified
	TimeLimit
)

const (
	ResourceTypeBag      = iota
	ResourceTypeCurrency // 货币
)

const (
	DisConnectMaxTimeLimit = 1 * time.Minute
)
