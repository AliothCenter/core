package stellar

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
	log "studio.sunist.work/platform/alioth-center/infrastructure/utils/logger"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"

	"gorm.io/gorm"

	"studio.sunist.work/platform/alioth-center/core/model"
	"studio.sunist.work/platform/alioth-center/infrastructure/database"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
)

var defaultDao *dao = nil

func init() {
	if initialize.GlobalConfig().Stellar.Storage == "postgres" || initialize.GlobalConfig().Stellar.Storage == "" {
		defaultDao = &dao{
			db:    database.GetGormAccessor(uint64(0), model.InstanceDTO{}, model.InstancePO{}),
			locks: map[string]*sync.RWMutex{},
		}

		if initialize.GlobalConfig().Stellar.Logger != "" {
			defaultDao.logger = log.NewLogger(initialize.GlobalConfig().Stellar.Logger)
		} else {
			defaultDao.logger = log.DefaultLogger()
		}

		if err := database.RegisterSyncModels(model.InstancePO{}); err != nil {
			defaultDao.logger.Log(log.DefaultField().WithLevel(log.Panic).WithCaller(log.Module).
				WithMessage("failed to register sync models").WithExtra(err.Error()))
		}
	}
}

func defaultDAO() *dao {
	return defaultDao
}

type dao struct {
	db     *database.GormAccessor[uint64, model.InstanceDTO, model.InstancePO]
	logger *log.Logger
	locks  map[string]*sync.RWMutex
}

// AddInstance 添加服务实例
func (d *dao) AddInstance(ctx context.Context, instanceService string, instanceVersion version.Version, ip string, port int) (dto model.InstanceDTO, err errors.AliothError) {
	// 如果没有锁，就创建一个
	if d.locks[instanceService] == nil {
		d.locks[instanceService] = &sync.RWMutex{}
	}

	// 装填服务内容
	instance := model.InstanceDTO{
		Address: fmt.Sprintf("%s:%d", ip, port),
		Service: instanceService,
		Version: instanceVersion.FormatDatabase(),
	}

	// 进行服务实例查询，此时不允许其他并发操作，因为可能导致实例编号不一致
	d.locks[instanceService].RLock()
	instanceIndex, countInstanceErr := d.db.CustomQueryCount("service = ? and version = ?", instance.Service, instance.Version)
	d.locks[instanceService].RUnlock()

	// 准备开始写入，为了防止此时有其他并发操作，加写锁
	d.locks[instanceService].Lock()
	if countInstanceErr != nil {
		if countInstanceErr.Derive(gorm.ErrRecordNotFound) {
			// 如果出现ErrRecordNotFound错误，说明一条记录都没有，那么实例编号就从0开始
			instanceIndex = 0
		} else {
			// 如果不是ErrRecordNotFound错误，说明出现了其他错误，返回错误信息
			d.locks[instanceService].Unlock()
			d.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar count instance error", ctx).
				WithExtraField("error", countInstanceErr.Error()).WithExtra(instance))
			return model.InstanceDTO{}, countInstanceErr
		}
	}

	// 如果实例编号大于等于48，说明实例编号已经用完了，返回错误信息
	if instanceIndex >= 48 {
		d.locks[instanceService].Unlock()
		d.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar instance out of max number", ctx, instance))
		return model.InstanceDTO{}, errors.NewInstanceOutOfMaxNumberError(48)
	}

	// 如果没有出现ErrRecordNotFound错误，说明至少有一条记录，那么实例编号就从记录数+1开始
	instanceIndex += 1
	// instance.Name: service:v0.0.0.1:alpha
	instanceName := strings.Builder{}
	instanceName.WriteString(instance.Service)
	instanceName.WriteString(":v")
	instanceName.WriteString(instanceVersion.Export())
	instanceName.WriteString(":")
	instanceName.WriteString(utils.GenerateGreeceAlphabetString(int(instanceIndex)))
	instance.Name = instanceName.String()

	// 执行插入操作
	if insertErr := d.db.InsertOne(instance); insertErr != nil {
		d.locks[instanceService].Unlock()
		d.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar insert instance error", ctx).
			WithExtraField("error", insertErr.Error()).WithExtra(instance))
		return model.InstanceDTO{}, insertErr
	} else {
		d.locks[instanceService].Unlock()
		d.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar insert instance success", ctx, instance))
		return instance, nil
	}
}

// RemoveInstance 删除服务实例
func (d *dao) RemoveInstance(ctx context.Context, serviceName, instanceName string) (err errors.AliothError) {
	// 如果没有锁，就创建一个
	if d.locks[serviceName] == nil {
		d.locks[serviceName] = &sync.RWMutex{}
	}

	d.locks[serviceName].RLock()
	_, queryErr := d.db.CustomQueryOne("name = ?", instanceName)
	d.locks[serviceName].RUnlock()
	if queryErr != nil {
		if queryErr.Derive(gorm.ErrRecordNotFound) {
			return errors.NewNoAvailableServiceError(serviceName, instanceName)
		} else {
			return queryErr
		}
	}

	// 进行服务实例删除，此时不允许其他并发操作，因为可能导致实例编号不一致
	d.locks[serviceName].Lock()
	deleteErr := d.db.DeleteOneByCondition(model.InstanceDTO{Name: instanceName})
	d.locks[serviceName].Unlock()
	if deleteErr != nil {
		d.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar delete instance error",
			ctx, model.InstanceDTO{Name: instanceName, Service: serviceName}).WithExtraField("error", deleteErr.Error()))
		return deleteErr
	} else {
		d.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar delete instance success",
			ctx, model.InstanceDTO{Name: instanceName, Service: serviceName}))
		return nil
	}
}

// FindInstance 查询服务实例
func (d *dao) FindInstance(ctx context.Context, service string, minVersion version.Version) (instances []model.InstanceDTO, err errors.AliothError) {
	// 如果没有锁，就创建一个
	if d.locks[service] == nil {
		d.locks[service] = &sync.RWMutex{}
	}

	// 进行服务实例查询，加读锁
	d.locks[service].RLock()
	if queryResult, queryErr := d.db.CustomQueryList("service = ? and version >= ?", service, minVersion); queryErr != nil {
		if queryErr.Derive(gorm.ErrRecordNotFound) {
			// 如果出现ErrRecordNotFound错误，说明一条记录都没有，直接返回空切片
			d.locks[service].RUnlock()
			d.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar find instance with no results",
				ctx, model.InstanceDTO{Service: service, Version: minVersion.FormatDatabase()}))
			return []model.InstanceDTO{}, nil
		} else {
			// 如果不是ErrRecordNotFound错误，说明出现了其他错误，返回错误信息
			d.locks[service].RUnlock()
			d.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar find instance error",
				ctx, model.InstanceDTO{Service: service, Version: minVersion.FormatDatabase()}).
				WithExtraField("error", queryErr.Error()))
			return []model.InstanceDTO{}, queryErr
		}
	} else {
		// 如果没有出现错误，直接返回结果
		d.locks[service].RUnlock()
		d.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar find instance success",
			ctx, model.InstanceDTO{Service: service, Version: minVersion.FormatDatabase()}).WithExtra("result", queryResult))
		return queryResult, nil
	}
}

func (d *dao) ListInstances(ctx context.Context, pageLimit, offset int) (instances []model.InstanceDTO, err error) {
	if queryResult, queryErr := d.db.CustomQueryListWithPaging(pageLimit, offset, "version >= ?", version.AlphaVersion.FormatDatabase()); queryErr != nil {
		if queryErr.Derive(gorm.ErrRegistered) {
			// 如果出现ErrRecordNotFound错误，说明一条记录都没有，直接返回空切片
			d.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar list instance with no results", ctx))
			return []model.InstanceDTO{}, nil
		} else {
			// 如果不是ErrRecordNotFound错误，说明出现了其他错误，返回错误信息
			d.logger.Log(log.DefaultField().WithFields(log.Error, log.Module, "alioth-stellar list instance error", ctx))
			return []model.InstanceDTO{}, queryErr
		}
	} else {
		// 如果没有出现错误，直接返回结果
		d.logger.Log(log.DefaultField().WithFields(log.Info, log.Module, "alioth-stellar list instance success", ctx))
		return queryResult, nil
	}
}
