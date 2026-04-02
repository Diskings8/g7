package dao_login

import (
	"g7/common/mysqlx"
	"g7/login/model_login"
)

// CreateUser 创建用户
func CreateUser(user *model_login.User) error {
	return mysqlx.GlobalDb.Create(user).Error
}

// GetUserByUsername 通过用户名找用户
func GetUserByUsername(username string) (*model_login.User, error) {
	var user model_login.User
	err := mysqlx.GlobalDb.Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetUserByID 通过用户id
func GetUserByID(userID int64) (*model_login.User, error) {
	var user model_login.User
	err := mysqlx.GlobalDb.First(&user, userID).Error
	return &user, err
}
