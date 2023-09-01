package restoration

import (
	"github.com/gin-gonic/gin"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

type HttpServer struct{}

func (h HttpServer) Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}

func (h HttpServer) Collection(ctx *gin.Context) {
	var request alioth.RestorationCollectionRequest
	if bindJsonErr := ctx.ShouldBindJSON(&request); bindJsonErr != nil {
		ctx.JSON(400, gin.H{
			"message": "invalid request",
			"error":   bindJsonErr.Error(),
		})
	} else {
		defaultService.CollectLogExternal(ctx, &request)
		ctx.JSON(200, gin.H{
			"message": "success",
		})
	}
}
