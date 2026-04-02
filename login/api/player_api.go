package api

import (
	"bytes"
	"encoding/json"
	"g7/common/jwt"
	"g7/login/internal/dao_login"
	"g7/login/internal/service_login"
	"g7/login/req_login"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

// PlayerList 获取角色列表
func PlayerList(c *gin.Context) {
	type Req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	players, err := service_login.ListPlayersByUserID(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取角色列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": players})
}

// CreatePlayer 创建角色
func CreatePlayer(c *gin.Context) {

	var req req_login.CreatePlayerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 1. 获取区服地址
	server, err := dao_login.GetServerByID(req.ServerID)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "区服不存在"})
		return
	}

	// 2. 转发请求到游戏服内部接口
	url := "http://" + server.Addr + "/api/game/create_player"
	body, _ := json.Marshal(map[string]any{
		"user_id":   req.UserID,
		"nickname":  req.Nickname,
		"server_id": req.ServerID,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "游戏服连接失败:" + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 3. 返回游戏服结果
	result, _ := io.ReadAll(resp.Body)
	c.Data(200, "application/json", result)
}

// SelectPlayer 选角（返回游戏服Token+地址）
func SelectPlayer(c *gin.Context) {

	var reqs req_login.SelectPlayerReq
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	//动态查询游戏服地址
	gServer, err := service_login.GetServerByID(reqs.ServerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "区服不存在或维护中"})
		return
	}

	player, err := service_login.SelectPlayer(reqs.UserID, reqs.UID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 生成游戏服Token（含账号ID+角色UID）
	gameToken, err := jwt.GenGameToken(reqs.UserID, reqs.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成游戏Token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":        200,
		"msg":         "选角成功",
		"game_token":  gameToken,
		"game_server": gServer.Addr, // 游戏服地址
		"player":      player,
	})
}
