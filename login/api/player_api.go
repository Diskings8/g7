package api

import (
	"context"
	"encoding/json"
	"g7/common/jwt"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/login/internal/dao_login"
	"g7/login/internal/service_login"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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

	var req pb.Req_Http_CreatePlayer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 1. 获取区服地址
	server, err := dao_login.GetServerByID(req.GetServerID())
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "区服不存在"})
		return
	}

	// 2. 转发请求到游戏服内部接口
	ctxConn, cancelConn := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelConn()
	client, err := protocol.NewGameNodeClient(ctxConn, server.Addr)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "注册服务暂不可用"})
		return
	}
	nodeReq := pb.Req_Node_CreatePlayer{
		UserID:   req.GetUserID(),
		ServerID: req.GetServerID(),
		Nickname: req.GetNickname(),
	}
	ctxReq, cancelReq := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelReq()
	nodeRsp, err := client.LoginNodeCreatePlayer(ctxReq, &nodeReq)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "游戏服连接失败:" + err.Error()})
		return
	}
	// 3. 返回游戏服结果
	rsp := pb.Rsp_Http_CreatePlayer{
		PlayerID: nodeRsp.PlayerID,
		ServerID: nodeRsp.ServerID,
		Nickname: nodeRsp.Nickname,
		ID:       nodeRsp.ID,
		UserID:   nodeRsp.UserID,
	}
	rsp.Token, _ = jwt.GenGameToken(rsp.GetUserID(), rsp.GetPlayerID(), rsp.GetServerID())

	result, _ := json.Marshal(&rsp)
	c.Data(200, "application/json", result)
}

// SelectPlayer 选角（返回游戏服Token+地址）
func SelectPlayer(c *gin.Context) {

	var reqs pb.Req_Http_SelectPlayer
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	//动态查询游戏服地址
	_, err := service_login.GetServerByID(reqs.ServerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "区服不存在或维护中"})
		return
	}

	_, err = service_login.SelectPlayer(reqs.GetUID(), reqs.GetPlayerID())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 生成游戏服Token（含账号ID+角色UID）
	gameToken, err := jwt.GenGameToken(reqs.GetUID(), reqs.GetPlayerID(), reqs.GetServerID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成游戏Token失败"})
		return
	}
	rsp := pb.Rsp_Http_SelectPlayer{
		ServerID: reqs.GetServerID(),
		PlayerID: reqs.GetPlayerID(),
		ID:       reqs.GetUID(),
		UserID:   reqs.GetUID(),
		Token:    gameToken,
	}

	c.JSON(http.StatusOK, &rsp)
}
