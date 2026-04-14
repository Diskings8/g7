package service_login

import (
	"errors"
	"g7/common/snowflakes"
	"g7/login/internal/dao_login"
	"g7/login/model_login"
	"golang.org/x/crypto/bcrypt"
)

// Register 账号注册
func (hts *loginHttpServer) Register(username, password string) error {
	// 检查用户名是否存在
	existUser, err := dao_login.GetUserByUsername(username)
	if err == nil && existUser != nil {
		return errors.New("用户名已存在")
	}

	// 加密密码
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建账号
	user := &model_login.User{
		ID:       snowflakes.GenUID(),
		Username: username,
		Password: string(hashPwd),
	}
	return dao_login.CreateUser(user)
}

// Login 账号登录（校验封禁）
func (hts *loginHttpServer) Login(username, password string) (*model_login.User, error) {
	user, err := dao_login.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 校验账号是否封禁
	if user.IsBand {
		return nil, errors.New("账号已被封禁")
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}
