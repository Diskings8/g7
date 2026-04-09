package main

import (
	"flag"
	"g7/common/config"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/utils"
	"github.com/gin-gonic/gin"
)

func main() {

	// 1. 解析环境参数
	flag.StringVar(&globals.Env, "env", "test", "运行环境: test/prod")
	flag.Parse()

	// 2、获取配置
	var confStr string
	if !utils.IsDev() {
		confStr = globals.ConfPro
	} else {
		confStr = globals.ConfDev
	}
	config.Load(confStr)

	logger.Init()

	r := gin.Default()

	// 平台服接口：区服列表、渠道配置、支付回调
	r.GET("/api/platform/server-list", func(c *gin.Context) {
		c.JSON(200, gin.H{"servers": []gin.H{{"id": 1, "name": "官方服-1"}}})
	})

	_ = r.Run(config.GCfg.Server.Platform)
}
