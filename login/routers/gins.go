package routers

import (
	"g7/login/api"

	"github.com/gin-gonic/gin"
)

var GEngine *gin.Engine

func GetDefaultGin() *gin.Engine {
	GEngine = gin.Default()
	return GEngine
}

func Register(r *gin.Engine) {

	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", api.Register)
		authGroup.POST("/login", api.Login)

	}
	playerGroup := r.Group("/api/player")
	{
		playerGroup.POST("/list", api.PlayerList)
		playerGroup.POST("/create", api.CreatePlayer)
		playerGroup.POST("/select", api.SelectPlayer)
	}

	orderGroup := r.Group("/api/order")
	{
		orderGroup.POST("/91/callback", api.GCallBack91.OrderCallBack)
	}
}
