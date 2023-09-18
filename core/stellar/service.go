package stellar

import (
	"context"
	"fmt"
	"math/rand"

	"studio.sunist.work/platform/alioth-center/infrastructure/global"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

var defaultService Service

func init() {
	if initialize.GlobalConfig().Stellar.Storage == "redis" {
		defaultService = &redisBasedService{
			redis: defaultCache(),
		}
	} else {
		defaultService = &postgresBasedService{
			dao: defaultDAO(),
		}
	}
}

type Service interface {
	ServiceRegistration(ctx context.Context, request *alioth.ServiceRegistrationRequest, ip string) (*alioth.ServiceRegistrationResponse, error)
	ServiceDiscovery(ctx context.Context, request *alioth.ServiceDiscoveryRequest) (*alioth.ServiceDiscoveryResponse, error)
	ServiceUnmount(ctx context.Context, request *alioth.ServiceUnmountRequest) (*alioth.ServiceUnmountResponse, error)
	ServiceList(ctx context.Context, request *alioth.ServiceListRequest) (*alioth.ServiceListResponse, error)
}

type postgresBasedService struct {
	dao *dao
}

func (s *postgresBasedService) ServiceRegistration(ctx context.Context, request *alioth.ServiceRegistrationRequest, ip string) (*alioth.ServiceRegistrationResponse, error) {
	versionFromExport, getVersionErr := version.NewVersionFromExport(request.GetVersion())
	if getVersionErr != nil {
		return nil, fmt.Errorf("failed to get version: %w", getVersionErr)
	}

	if result, addInstanceErr := s.dao.AddInstance(ctx, request.GetService(), versionFromExport, ip, int(request.GetPort())); addInstanceErr != nil {
		return nil, fmt.Errorf("failed to add instance: %w", addInstanceErr)
	} else {
		return &alioth.ServiceRegistrationResponse{
			Service: result.Service,
			Address: result.Address,
			Name:    result.Name,
			Version: version.Version(result.Version).Export(),
		}, nil
	}
}

func (s *postgresBasedService) ServiceDiscovery(ctx context.Context, request *alioth.ServiceDiscoveryRequest) (*alioth.ServiceDiscoveryResponse, error) {
	minVersion, getVersionErr := version.NewVersionFromExport(request.GetMinVersion())
	if getVersionErr != nil {
		return nil, fmt.Errorf("failed to get min version: %w", getVersionErr)
	}

	if instances, getInstancesErr := s.dao.FindInstance(ctx, request.GetService(), minVersion); getInstancesErr != nil {
		return nil, fmt.Errorf("failed to find instance: %w", getInstancesErr)
	} else if len(instances) == 0 {
		// 如果没有找到实例，返回空
		return nil, errors.NewNoAvailableInstanceError(request.GetService(), request.GetMinVersion())
	} else {
		// 使用随机负载均衡
		instance := instances[rand.Intn(len(instances))]
		return &alioth.ServiceDiscoveryResponse{
			Service:     instance.Service,
			Name:        instance.Name,
			Version:     version.Version(instance.Version).Export(),
			Address:     instance.Address,
			LastUpdated: instance.UpdatedAt.Format(global.AliothTimeFormat),
		}, nil
	}
}

func (s *postgresBasedService) ServiceUnmount(ctx context.Context, request *alioth.ServiceUnmountRequest) (*alioth.ServiceUnmountResponse, error) {
	if removeInstanceErr := s.dao.RemoveInstance(ctx, request.GetService(), request.GetName()); removeInstanceErr != nil {
		return nil, fmt.Errorf("failed to remove instance: %w", removeInstanceErr)
	} else {
		return &alioth.ServiceUnmountResponse{
			Service: request.GetService(),
			Name:    request.GetName(),
			Success: true,
		}, nil
	}
}

func (s *postgresBasedService) ServiceList(ctx context.Context, request *alioth.ServiceListRequest) (*alioth.ServiceListResponse, error) {
	if instances, getInstancesErr := s.dao.ListInstances(ctx, int(request.GetPageLimit()), int(request.GetPageOffset())); getInstancesErr != nil {
		return nil, fmt.Errorf("failed to list instance: %w", getInstancesErr)
	} else {
		list := make([]*alioth.ServiceRecord, len(instances))
		for i, instance := range instances {
			list[i] = &alioth.ServiceRecord{
				Service:   instance.Service,
				Name:      instance.Name,
				Address:   instance.Address,
				Version:   version.Version(instance.Version).Export(),
				UpdatedAt: instance.UpdatedAt.Format(global.AliothTimeFormat),
				CreatedAt: instance.CreatedAt.Format(global.AliothTimeFormat),
			}
		}
		return &alioth.ServiceListResponse{
			Total:      int32(len(instances)),
			PageLimit:  request.GetPageLimit(),
			PageOffset: request.GetPageOffset(),
			Services:   list,
		}, nil
	}
}

type redisBasedService struct {
	redis *cache
}

func (r *redisBasedService) ServiceRegistration(ctx context.Context, request *alioth.ServiceRegistrationRequest, ip string) (*alioth.ServiceRegistrationResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (r *redisBasedService) ServiceDiscovery(ctx context.Context, request *alioth.ServiceDiscoveryRequest) (*alioth.ServiceDiscoveryResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (r *redisBasedService) ServiceUnmount(ctx context.Context, request *alioth.ServiceUnmountRequest) (*alioth.ServiceUnmountResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (r *redisBasedService) ServiceList(ctx context.Context, request *alioth.ServiceListRequest) (*alioth.ServiceListResponse, error) {
	// TODO implement me
	panic("implement me")
}
