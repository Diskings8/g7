package api

import (
	"encoding/json"
	"g7/game/req_game"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreatePlayer 创建角色
func CreatePlayer(c *gin.Context) {

	var req req_game.CreatePlayerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 2. 转发请求到游戏服内部接口

	// 3. 返回游戏服结果
	result := struct {
		ID       int64
		UID      int64
		UserID   int64
		ServerID int64
		Nickname string
	}{
		ID:       100001,
		UID:      910001,
		UserID:   req.UserID,
		ServerID: int64(req.ServerID),
		Nickname: req.Nickname,
	}
	H, _ := json.Marshal(result)

	c.Data(200, "application/json", H)
}
