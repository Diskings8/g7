package routers

import (
	"g7/game/api"

	"github.com/gin-gonic/gin"
)

var GEngine *gin.Engine

func GetDefaultGin() *gin.Engine {
	GEngine = gin.Default()
	return GEngine
}

func Register(r *gin.Engine) {

	gGroup := r.Group("/api/game")
	{
		gGroup.POST("/create_player", api.CreatePlayer)
	}
}
