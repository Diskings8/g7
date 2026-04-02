package main

import (
	"g7/pkg/logger"
	"g7/pkg/mysqlx"
	"g7/pkg/redisx"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化日志
	logger.Init()

	// 2. 初始化 MySQL
	mysqlx.Init("root:@tcp(127.0.0.1:3306)/game_db?charset=utf8mb4&parseTime=True&loc=Local")

	// 3. 初始化 Redis
	redisx.Init("127.0.0.1:6379", "", 0)

	// 4. 启动服务
	r := gin.Default()

	logger.Log.Info("服务启动成功 :8080")
	r.Run(":8080")
}
