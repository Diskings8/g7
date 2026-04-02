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

	authGroup := r.Group("/api/login")
	{
		authGroup.POST("/register", api.Register)
		authGroup.POST("/accountLogin", api.Login)
		authGroup.POST("/player/list", api.PlayerList)
		authGroup.POST("/create", api.CreatePlayer)
		authGroup.POST("/select", api.SelectPlayer)
	}
}
