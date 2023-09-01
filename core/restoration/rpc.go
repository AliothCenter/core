package restoration

import (
	"context"

	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

type RpcServer struct {
	alioth.UnimplementedAliothRestorationServer
}

func (a RpcServer) RestorationCollection(ctx context.Context, request *alioth.RestorationCollectionRequest) (*alioth.RestorationCollectionResponse, error) {
	defaultService.CollectLog(ctx, request)
	return &alioth.RestorationCollectionResponse{}, nil
}
