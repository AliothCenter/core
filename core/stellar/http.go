package stellar

import (
	"strings"

	"github.com/gin-gonic/gin"

	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

type HttpServer struct{}

func (h HttpServer) Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}

func (h HttpServer) ServiceRegistration(ctx *gin.Context) {
	request := alioth.ServiceRegistrationRequest{}
	if bindJsonErr := ctx.ShouldBind(&request); bindJsonErr != nil {
		ctx.JSON(400, gin.H{
			"message": "invalid request",
			"error":   bindJsonErr.Error(),
		})
	} else if response, registrationErr := defaultService.ServiceRegistration(ctx, &request, ctx.RemoteIP()); registrationErr != nil {
		ctx.JSON(500, gin.H{
			"message": "internal error",
			"error":   registrationErr.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{
			"message": "success",
			"data":    response,
		})
	}
}

func (h HttpServer) ServiceDiscovery(ctx *gin.Context) {
	request := alioth.ServiceDiscoveryRequest{}
	serviceName, minVersion := ctx.Param("service"), ctx.Query("min_version")
	if serviceName == "" || minVersion == "" || len(strings.Split(minVersion, ".")) != 4 {
		ctx.JSON(400, gin.H{
			"message": "invalid request",
			"error":   "invalid service name or min version",
		})
		return
	} else {
		request.Service = serviceName
		request.MinVersion = minVersion
	}

	if response, discoveryErr := defaultService.ServiceDiscovery(ctx, &request); discoveryErr != nil {
		ctx.JSON(500, gin.H{
			"message": "internal error",
			"error":   discoveryErr.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{
			"message": "success",
			"data":    response,
		})
	}
}

func (h HttpServer) ServiceUnmount(ctx *gin.Context) {
	request := alioth.ServiceUnmountRequest{}
	serviceName, handlerName := ctx.Param("service"), ctx.Param("handler")
	if serviceName == "" || handlerName == "" || len(strings.Split(handlerName, ":")) != 3 {
		ctx.JSON(400, gin.H{
			"message": "invalid request",
			"error":   "invalid service name or handler name",
		})
		return
	} else {
		request.Service = serviceName
		request.Name = handlerName
	}

	if response, unmountErr := defaultService.ServiceUnmount(ctx, &request); unmountErr != nil {
		ctx.JSON(500, gin.H{
			"message": "internal error",
			"error":   unmountErr.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{
			"message": "success",
			"data":    response,
		})
	}
}

func (h HttpServer) ServiceList(ctx *gin.Context) {
	request := alioth.ServiceListRequest{}
	if bindJsonErr := ctx.ShouldBind(&request); bindJsonErr != nil {
		ctx.JSON(400, gin.H{
			"message": "invalid request",
			"error":   bindJsonErr.Error(),
		})
	} else if response, listErr := defaultService.ServiceList(ctx, &request); listErr != nil {
		ctx.JSON(500, gin.H{
			"message": "internal error",
			"error":   listErr.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{
			"message": "success",
			"data":    response,
		})
	}
}
