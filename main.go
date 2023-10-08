package main

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"studio.sunist.work/platform/alioth-center/core/restoration"
	"studio.sunist.work/platform/alioth-center/core/stellar"
	"studio.sunist.work/platform/alioth-center/infrastructure/database"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	log "studio.sunist.work/platform/alioth-center/infrastructure/utils/logger"
)

func main() {
	// 初始化数据库
	database.SyncDatabase()

	// 初始化rpc和http服务器
	logger := log.DefaultLogger()
	lis, err := net.Listen("tcp",
		fmt.Sprintf("%s:%d", initialize.GlobalConfig().Grpc.ListenIP, initialize.GlobalConfig().Grpc.ListenPort))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	engine := gin.Default()
	engine.Use(gin.Recovery())
	external := engine.Group("/external")

	// 注册rpc和http服务器
	restoration.InitRestorationRpcServer(s)
	restoration.InitRestorationHttpServer(external)
	stellar.InitStellarRpcServer(s)
	stellar.InitStellarHttpServer(external)

	// 启动rpc和http服务器
	rpcExit := startRpcEngine(s, lis)
	httpExit := startHttpEngine(engine,
		fmt.Sprintf("%s:%d", initialize.GlobalConfig().Http.ListenIP, initialize.GlobalConfig().Http.ListenPort))
	logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Info).WithMessage("server(s) started"))

	// 注册服务到stellar
	restoration.RegisterToStellar()

	// 等待rpc和http服务器退出
	select {
	case err := <-rpcExit:
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("rpc server exit").WithExtra(err.Error()))
	case err := <-httpExit:
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("rpc server exit").WithExtra(err.Error()))
	}
}

func startHttpEngine(engine *gin.Engine, addr string) (exitChan chan error) {
	exit := make(chan error)
	go func() {
		exit <- engine.Run(addr)
	}()
	return exit
}

func startRpcEngine(engine *grpc.Server, conn net.Listener) (exitChan chan error) {
	exit := make(chan error)
	go func() {
		exit <- engine.Serve(conn)
	}()
	return exit
}
