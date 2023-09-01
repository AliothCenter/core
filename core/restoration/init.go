package restoration

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

func InitRestorationRpcServer(server *grpc.Server) {
	alioth.RegisterAliothRestorationServer(server, &RpcServer{})
}

func InitRestorationHttpServer(group *gin.RouterGroup) {
	server := HttpServer{}
	group.GET("/restoration/ping", server.Ping)
	group.POST("/restoration/collection", server.Collection)
}
