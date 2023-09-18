package stellar

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"studio.sunist.work/platform/alioth-center/core/model"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	"studio.sunist.work/platform/alioth-center/infrastructure/memcache"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
	log "studio.sunist.work/platform/alioth-center/infrastructure/utils/logger"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
)

var redisCache *cache = nil

func init() {
	if initialize.GlobalConfig().Stellar.Storage == "redis" {
		redisCache = &cache{
			client: memcache.GetRedisCache(),
		}

		if initialize.GlobalConfig().Stellar.Logger != "" {
			defaultDao.logger = log.NewLogger(initialize.GlobalConfig().Stellar.Logger)
		} else {
			defaultDao.logger = log.DefaultLogger()
		}
	}
}

func defaultCache() *cache {
	return redisCache
}

type cache struct {
	client *redis.Client
	logger *log.Logger
}

func (c *cache) AddInstance(ctx context.Context, instanceService string, instanceVersion version.Version, ip string, port int) (dto model.InstanceDTO, err errors.AliothError) {
	// 装填服务内容
	instance := model.InstanceDTO{
		Address:   fmt.Sprintf("%s:%d", ip, port),
		Service:   instanceService,
		Version:   instanceVersion.FormatDatabase(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置服务版本
	if c.client.Exists(ctx, utils.BuildRedisKey(instanceService)).Val() == 0 {
		// 如果这个服务不存在，创建并将服务版本添加到 set 中
		c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar insert instance version success", ctx, instanceVersion.Format()))
		c.client.SAdd(ctx, utils.BuildRedisKey(instanceService, "versions"), instanceVersion.Export())
	} else if !c.client.SIsMember(ctx, utils.BuildRedisKey(instanceService, "versions"), instanceVersion.Export()).Val() {
		// 如果这个服务存在，但是版本不存在，将服务版本添加到 set 中
		c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar insert instance version success", ctx, instanceVersion.Format()))
		c.client.SAdd(ctx, utils.BuildRedisKey(instanceService, "versions"), instanceVersion.Export())
	}

	// 设置服务信息
	if c.client.Exists(ctx, utils.BuildRedisKey(instanceService, instanceVersion.Export())).Val() == 0 {
		instanceName := strings.Builder{}
		instanceName.WriteString(instance.Service)
		instanceName.WriteString(":v")
		instanceName.WriteString(instanceVersion.Export())
		instanceName.WriteString(":")
		instanceName.WriteString(utils.GenerateGreeceAlphabetString(1))
		instance.Name = instanceName.String()
		marshalBytes := utils.JsonMarshal(instance)
		c.client.SAdd(ctx, utils.BuildRedisKey(instanceService, instanceVersion.Export()), string(marshalBytes))
		c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar insert instance success", ctx, instance))
		return instance, nil
	}

	// 存在这个服务，需要命名，然后创建
	if instanceIndex, countInstanceErr := c.client.SCard(ctx, utils.BuildRedisKey(instanceService, instanceVersion.Export())).Result(); countInstanceErr != nil {
		c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar insert instance error", ctx).WithExtraField("error", countInstanceErr.Error()).WithExtra(instance))
		return model.InstanceDTO{}, errors.NewExecuteSqlError("SCard", countInstanceErr)
	} else if instanceIndex >= 48 {
		c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar instance out of max number", ctx, instance))
		return model.InstanceDTO{}, errors.NewInstanceOutOfMaxNumberError(48)
	} else {
		instanceName := strings.Builder{}
		instanceName.WriteString(instance.Service)
		instanceName.WriteString(":v")
		instanceName.WriteString(instanceVersion.Export())
		instanceName.WriteString(":")
		instanceName.WriteString(utils.GenerateGreeceAlphabetString(int(instanceIndex + 1)))
		instance.Name = instanceName.String()
		marshalBytes := utils.JsonMarshal(instance)
		c.client.SAdd(ctx, utils.BuildRedisKey(instanceService, instanceVersion.Export()), string(marshalBytes))
		c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar insert instance success", ctx, instance))
		return instance, nil
	}
}

func (c *cache) RemoveInstance(ctx context.Context, serviceName, instanceName string) (err errors.AliothError) {
	// 获取服务版本
	instanceVersion, getVersion := version.NewVersionFromInstanceName(instanceName)
	if getVersion != nil {
		return errors.NewNoAvailableServiceError(serviceName, instanceName)
	}

	// 检查服务是否存在
	if c.client.Exists(ctx, utils.BuildRedisKey(serviceName, instanceVersion.Export())).Val() == 0 {
		return errors.NewNoAvailableServiceError(serviceName, instanceName)
	}

	//  删除服务
	if removeInstanceErr := c.client.SRem(ctx, utils.BuildRedisKey(serviceName, instanceVersion.Export())).Err(); removeInstanceErr != nil {
		c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar remove instance failed", ctx, removeInstanceErr.Error()))
		return errors.NewExecuteSqlError("SRem", removeInstanceErr)
	}

	// 删除服务版本
	if c.client.SCard(ctx, utils.BuildRedisKey(serviceName, instanceVersion.Export())).Val() == 0 {
		c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar remove instance version success", ctx, instanceVersion.Format()))
		c.client.SRem(ctx, utils.BuildRedisKey(serviceName, "versions"), instanceVersion.Export())
	}

	c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar remove instance success", ctx, instanceName))
	return nil
}

func (c *cache) FindInstance(ctx context.Context, service string, minVersion version.Version) (instances []model.InstanceDTO, err errors.AliothError) {
	if c.client.Exists(ctx, utils.BuildRedisKey(service, "versions")).Val() == 0 {
		c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar find instance with no results", ctx, model.InstanceDTO{Service: service, Version: minVersion.FormatDatabase()}))
		return []model.InstanceDTO{}, errors.NewNoAvailableServiceError(service, minVersion.Export())
	} else if versionStrings, getVersionsErr := c.client.SMembers(ctx, utils.BuildRedisKey(service, "versions")).Result(); getVersionsErr != nil {
		c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar find instance error", ctx, getVersionsErr.Error()))
		return []model.InstanceDTO{}, errors.NewExecuteSqlError("SMembers", getVersionsErr)
	} else {
		// 获取所有符合版本
		var versions []version.Version
		for _, versionString := range versionStrings {
			if v, gv := version.NewVersionFromExport(versionString); gv != nil {
				c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar get instance version error", ctx, gv.Error()))
			} else if v.FormatDatabase() >= minVersion.FormatDatabase() {
				versions = append(versions, v)
			}
		}

		// 如果没有符合版本，返回空
		if len(versions) == 0 {
			c.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar find instance with no results", ctx, model.InstanceDTO{Service: service, Version: minVersion.FormatDatabase()}))
			return []model.InstanceDTO{}, errors.NewNoAvailableServiceError(service, minVersion.Export())
		}

		// 获取所有符合版本的实例
		var instanceList []model.InstanceDTO
		for _, v := range versions {
			if instanceStrings, getInstanceErr := c.client.SMembers(ctx, utils.BuildRedisKey(service, v.Export())).Result(); getInstanceErr != nil {
				c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar find instance error", ctx, getInstanceErr.Error()))
				return []model.InstanceDTO{}, errors.NewExecuteSqlError("SMembers", getInstanceErr)
			} else {
				for _, instanceString := range instanceStrings {
					var instance model.InstanceDTO
					if unmarshalErr := json.Unmarshal([]byte(instanceString), &instance); unmarshalErr != nil {
						c.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar unmarshal instance error", ctx, unmarshalErr.Error()))
						return []model.InstanceDTO{}, errors.NewExecuteSqlError("JsonUnmarshal", unmarshalErr)
					} else {
						instanceList = append(instanceList, instance)
					}
				}
			}
		}

		return instances, nil
	}
}
