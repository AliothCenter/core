package stellar

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

func InitStellarRpcServer(server *grpc.Server) {
	alioth.RegisterAliothStellarServer(server, &RpcServer{})
}

func InitStellarHttpServer(group *gin.RouterGroup) {
	server := HttpServer{}
	group.GET("/stellar/ping", server.Ping)
	group.POST("/stellar/registration", server.ServiceRegistration)
	group.GET("/stellar/discovery/:service", server.ServiceDiscovery)
	group.DELETE("/stellar/unmount/:service/:handler", server.ServiceUnmount)
	group.GET("/stellar/list", server.ServiceList)
}
