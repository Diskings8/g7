package api

import (
	"context"
	"fmt"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/login/global_login"
	"g7/login/model_login"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var GCallBack91 callBack_91

type callBack_91 struct {
}

// OrderCallBack
func (this *callBack_91) OrderCallBack(c *gin.Context) {

	var req model_login.PaymentCallBackReq
	var rsp = &model_login.PaymentCallBackRsp{
		Code: 400,
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.Msg = "参数错误"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	//校验
	order := &model_common.GameOrder{}
	_ = global_login.GLoginDB.FindOne(order, map[string]interface{}{"order_no": req.OrderId})

	// 存储回调信息，确认回调已到达
	orderPayment := model_common.PaymentRecord{}
	_ = global_login.GLoginDB.Insert(&orderPayment)

	// 校验合法性
	if !this.verifySign(req) {

		rsp.Msg = "签名校验失败"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	// 避免重复处理
	if order.Status == globals.OrderStatusPending {
		order.Status = globals.OrderStatusPaid
		order.PayAmount = req.PayAmount
		order.PayTime = req.PayTime
		order.PayType = req.PayType
		order.Currency = req.Currency
		_ = global_login.GLoginDB.Insert(order)
	} else {

		rsp.Msg = "订单已处理"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}
	key := fmt.Sprintf(redisx.PlayerLockKeyPrefix, order.ServerID, order.PlayerID)
	val, err := redisx.GetClient().Get(context.Background(), key).Result()
	if err != nil {
		logger.Log.Info(fmt.Sprintf("%s 服玩家不在线:%d, 订单未处理:%s", order.ServerID, order.PlayerID, order.OrderNo))
		rsp.Msg = "订单发货中"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	ctxConn, cancelConn := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelConn()
	client, err := protocol.NewGameNodeClient(ctxConn, val)
	if err != nil {
		rsp.Msg = "服务暂不可用"
		c.JSON(http.StatusBadRequest, rsp)
		return
	}
	nodeReq := pb.Req_Node_OrderPaid{
		PlayerID: order.PlayerID,
		ServerID: order.ServerID,
		OrderId:  order.OrderNo,
	}
	ctxReq, cancelReq := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelReq()
	_, err = client.LoginNodeOrderPaid(ctxReq, &nodeReq)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "游戏服连接失败:" + err.Error()})
		return
	}
	rsp.Msg = "服务暂不可用"
	c.JSON(http.StatusBadRequest, rsp)
}

func (this *callBack_91) verifySign(req model_login.PaymentCallBackReq) bool {
	return true
}
