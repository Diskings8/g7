package dao_login

import (
	"g7/login/global_login"
	"g7/login/model_login"
)

// CreateUser 创建用户
func CreateUser(user *model_login.User) error {
	return global_login.GLoginDB.Insert(user)
}

// GetUserByUsername 通过用户名找用户
func GetUserByUsername(username string) (*model_login.User, error) {
	var user model_login.User
	err := global_login.GLoginDB.FindOne(&user, map[string]any{"username": username})
	return &user, err
}

// GetUserByID 通过用户id
func GetUserByID(userID int64) (*model_login.User, error) {
	var user model_login.User
	err := global_login.GLoginDB.FindOne(&user, map[string]any{"user_Id": userID})
	return &user, err
}
