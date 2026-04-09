package mqc

import "fmt"

const (
	MQGameLog = "mqGame"
)

func MakeGameCreateRoleTopicKey() string {
	return fmt.Sprintf("%s_createRole", MQGameLog)
}

func MakeGameActionTopicKey() string {
	return fmt.Sprintf("%s_action", MQGameLog)
}
