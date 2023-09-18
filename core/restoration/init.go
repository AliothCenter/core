package restoration

import (
	"fmt"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/exit"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	stellar "studio.sunist.work/platform/alioth-center/core/stellar/client"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

func RegisterToStellar() {
	grpcConf := initialize.GlobalConfig().Grpc
	client, err := stellar.NewClient(fmt.Sprintf("%s:%d", grpcConf.ListenIP, grpcConf.ListenPort))
	if err != nil {
		panic(err)
	}

	_, handlerName, registerErr := client.Register("alioth-restoration", version.NewVersion(1, 0, 0, 0), grpcConf.ListenPort)
	if registerErr != nil {
		panic(registerErr)
	}

	exit.AddExitFunctions(func() error {
		err := client.Unmount("alioth-restoration", handlerName)
		if err != nil {
			return err
		}
		return nil
	})
}

func InitRestorationRpcServer(server *grpc.Server) {
	alioth.RegisterAliothRestorationServer(server, &RpcServer{})
}

func InitRestorationHttpServer(group *gin.RouterGroup) {
	server := HttpServer{}
	group.GET("/restoration/ping", server.Ping)
	group.POST("/restoration/collection", server.Collection)
}
