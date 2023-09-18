package restoration

import (
	"context"

	"studio.sunist.work/platform/alioth-center/infrastructure/utils/logger"

	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

var defaultService *Service

func init() {
	defaultService = &Service{
		logger: log.NewLogger("logs/restoration"),
	}
}

type Service struct {
	logger *log.Logger
}

func (s *Service) CollectLog(ctx context.Context, request *alioth.RestorationCollectionRequest) {
	s.logger.Log(NewRestorationFieldsFromRequest(ctx, request))
}

func (s *Service) CollectLogExternal(ctx context.Context, request *alioth.RestorationCollectionRequest) {
	s.logger.Log(NewExternalRestorationFieldsFromRequest(ctx, request))
}
