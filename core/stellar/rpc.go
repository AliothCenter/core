package stellar

import (
	"context"

	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

type RpcServer struct {
	alioth.UnimplementedAliothStellarServer
}

func (r RpcServer) ServiceRegistration(ctx context.Context, request *alioth.ServiceRegistrationRequest) (*alioth.ServiceRegistrationResponse, error) {
	if ip, getIPErr := utils.GetContextClientIP(ctx); getIPErr != nil {
		return nil, getIPErr
	} else {
		return defaultService.ServiceRegistration(ctx, request, ip)
	}
}

func (r RpcServer) ServiceDiscovery(ctx context.Context, request *alioth.ServiceDiscoveryRequest) (*alioth.ServiceDiscoveryResponse, error) {
	return defaultService.ServiceDiscovery(ctx, request)
}

func (r RpcServer) ServiceUnmount(ctx context.Context, request *alioth.ServiceUnmountRequest) (*alioth.ServiceUnmountResponse, error) {
	return defaultService.ServiceUnmount(ctx, request)
}

func (r RpcServer) ServiceList(ctx context.Context, request *alioth.ServiceListRequest) (*alioth.ServiceListResponse, error) {
	return defaultService.ServiceList(ctx, request)
}
