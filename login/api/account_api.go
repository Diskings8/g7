package api

import (
	"g7/common/configx"
	"g7/common/jwt"
	"g7/login/internal/service_login"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Register 注册接口
func Register(c *gin.Context) {
	if !configx.GEtcdCfg.RegisterOn {
		c.JSON(http.StatusBadRequest, gin.H{"code": 502, "msg": "已关闭注册"})
		return
	}

	type Req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if err := service_login.Register(req.Username, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "注册成功"})
}

// Login 登录接口（返回账号信息+角色列表）
func Login(c *gin.Context) {
	if !configx.GEtcdCfg.LoginOn {
		c.JSON(http.StatusBadRequest, gin.H{"code": 502, "msg": "已暂停登录"})
		return
	}
	type Req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	user, err := service_login.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 获取该账号下角色列表
	players, err := service_login.ListPlayersByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取角色列表失败"})
		return
	}

	// 生成登录Token（仅含账号ID）
	token, err := jwt.GenLoginToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成Token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"msg":     "登录成功",
		"token":   token,
		"user_id": user.ID,
		"players": players,
	})
}
