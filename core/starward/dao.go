package starward

import (
	"context"
	"time"

	"gorm.io/gorm"
	"studio.sunist.work/platform/alioth-center/core/model"
	"studio.sunist.work/platform/alioth-center/infrastructure/database"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	"studio.sunist.work/platform/alioth-center/infrastructure/memcache"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
	log "studio.sunist.work/platform/alioth-center/infrastructure/utils/logger"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/passwd"
)

var (
	defaultDao *dao = nil

	redisCallbackMessageKeyPrefix = "callback_message"
	redisCode2TokenKeyPrefix      = "code_to_token"
	redisTokenScopesKeyPrefix     = "token_scopes"
	redisUserInfoScopeKeyPrefix   = "user_info_scope"

	activateCodeTTL = time.Minute * 10
)

func init() {
	// 设置包内变量
	if activateCodeTTLConf := time.Duration(initialize.GlobalConfig().Starward.Security.ActivateCodeTimeoutMinutes) * time.Minute; activateCodeTTLConf != 0 {
		activateCodeTTL = activateCodeTTLConf
	}
	if redisCallbackMessageKeyPrefixConf := initialize.GlobalConfig().Starward.Keys.CallbackMessageKeyPrefix; redisCallbackMessageKeyPrefixConf != "" {
		redisCallbackMessageKeyPrefix = redisCallbackMessageKeyPrefixConf
	}
	if redisCode2TokenKeyPrefixConf := initialize.GlobalConfig().Starward.Keys.CodeToTokenKeyPrefix; redisCode2TokenKeyPrefixConf != "" {
		redisCode2TokenKeyPrefix = redisCode2TokenKeyPrefixConf
	}
	if redisTokenScopesKeyPrefixConf := initialize.GlobalConfig().Starward.Keys.TokenScopeKeyPrefix; redisTokenScopesKeyPrefixConf != "" {
		redisTokenScopesKeyPrefix = redisTokenScopesKeyPrefixConf
	}
	if redisUserInfoScopeKeyPrefixConf := initialize.GlobalConfig().Starward.Keys.UserInfoScopeKeyPrefix; redisUserInfoScopeKeyPrefixConf != "" {
		redisUserInfoScopeKeyPrefix = redisUserInfoScopeKeyPrefixConf
	}

	// 初始化默认dao
	defaultDao = &dao{
		db:           database.GetGorm(),
		users:        database.GetGormAccessor(uint64(0), model.UserDTO{}, model.User{}),
		applications: database.GetGormAccessor(uint64(0), model.ApplicationDTO{}, model.Application{}),
		scopes:       database.GetGormAccessor(uint64(0), model.ScopeDTO{}, model.Scope{}),
		rds:          memcache.GetRedisAccessor(initialize.GlobalConfig().Starward.Keys.ModuleRedisPrefix),
		logger:       log.DefaultLogger(),
	}
	if err := database.RegisterSyncModels(model.User{}, model.Application{}, model.Scope{}); err != nil {
		defaultDao.logger.Log(log.DefaultField().WithLevel(log.Panic).WithCaller(log.Module).WithMessage("failed to register sync models").WithExtra(err.Error()))
	}
}

type dao struct {
	db           *gorm.DB
	users        *database.GormAccessor[uint64, model.UserDTO, model.User]
	applications *database.GormAccessor[uint64, model.ApplicationDTO, model.Application]
	scopes       *database.GormAccessor[uint64, model.ScopeDTO, model.Scope]
	rds          *memcache.RedisAccessor
	logger       *log.Logger
}

// DeleteCode 删除验证码
//   - key: 验证码键
func (d *dao) DeleteCode(ctx context.Context, key string) (success bool, err errors.AliothError) {
	if deleteErr := d.rds.Delete(ctx, key); deleteErr != nil {
		// 删除验证码失败
		d.logger.LogModuleError(ctx, "failed to delete code", deleteErr)
		return false, deleteErr
	} else {
		// 删除验证码成功
		d.logger.LogModuleInfo(ctx, "delete code success", key)
		return true, nil
	}
}

// StoreCode 存储验证码
//   - key: 验证码键
//   - code: 验证码
func (d *dao) StoreCode(ctx context.Context, key string, code string) (err errors.AliothError) {
	if storeErr := d.rds.StoreEX(ctx, key, code, activateCodeTTL); storeErr != nil {
		return storeErr
	} else {
		return nil
	}
}

// CreateUser 创建用户
//   - username: 用户名
//   - password: 密码
//   - email: 邮箱
func (d *dao) CreateUser(ctx context.Context, username, password, email string) (user model.UserDTO, err errors.AliothError) {
	// 装填用户信息
	user = model.UserDTO{
		Username: username,
		Password: password,
		Email:    email,
		Role:     RoleNormal.ExportDatabase(),
		Enable:   false,
	}

	// 创建用户
	if createErr := d.users.InsertOne(user); createErr != nil {
		// 创建用户失败
		d.logger.LogModuleError(ctx, "failed to create user", createErr)
		return model.UserDTO{}, errors.Derive("failed to create user", createErr)
	} else {
		// 创建用户成功
		d.logger.LogModuleInfo(ctx, "create user success", username)
		return user, nil
	}
}

// CheckUserStatus 检查用户是否存在以及是否激活(邮箱)
//   - username: 用户名
func (d *dao) CheckUserStatus(ctx context.Context, username string) (exist bool, activated bool, err errors.AliothError) {
	// 查询用户是否激活
	if user, queryErr := d.users.CustomQueryOne("username = ?", username); queryErr == nil {
		// 查询用户成功
		d.logger.LogModuleInfo(ctx, "query user success", username)
		return true, user.Enable, nil
	} else if errors.Is(gorm.ErrRecordNotFound, queryErr) {
		// 用户不存在
		d.logger.LogModuleInfo(ctx, "user not exist", username)
		return false, false, nil
	} else {
		// 查询用户失败
		d.logger.LogModuleError(ctx, "failed to query user", queryErr)
		return false, false, errors.Derive("failed to query user", queryErr)
	}
}

// CheckEmailInuse 检查邮箱是否被使用
//   - email: 邮箱
func (d *dao) CheckEmailInuse(ctx context.Context, email string) (inuse bool, err errors.AliothError) {
	if count, countErr := d.users.CustomQueryCount("email = ?", email); countErr != nil {
		// 查询使用邮箱的用户数量失败
		d.logger.LogModuleError(ctx, "failed to count user", countErr)
		return false, errors.Derive("failed to count user", countErr)
	} else {
		// 查询使用邮箱的用户数量成功
		d.logger.LogModuleInfo(ctx, "count user success", email)
		return count != 0, nil
	}
}

// CheckUserEmail 检查用户邮箱是否匹配
//   - username: 用户名
//   - email: 邮箱
func (d *dao) CheckUserEmail(ctx context.Context, username, email string, activated bool) (correct bool, err errors.AliothError) {
	if queryErr := d.db.Model(&model.User{}).Where("username = ? and email = ? and enable = ?", username, email, activated).First(&model.User{}).Error; queryErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, queryErr) {
			d.logger.LogModuleInfo(ctx, "user not exist", username)
			return false, nil
		} else {
			d.logger.LogModuleError(ctx, "failed to query user", queryErr)
			return false, errors.Derive("failed to query user", queryErr)
		}
	} else {
		d.logger.LogModuleInfo(ctx, "query user success", username)
		return true, nil
	}
}

// CheckUserPassword 检查用户密码是否匹配
//   - username: 用户名
//   - password: 密码
func (d *dao) CheckUserPassword(ctx context.Context, username, password string) (correct bool, err errors.AliothError) {
	if user, queryErr := d.users.CustomQueryOne("username = ?", username); queryErr == nil {
		if !passwd.BcryptCheck(user.Password, password) {
			// 密码不匹配
			d.logger.LogModuleInfo(ctx, "password not match", username)
			return false, nil
		} else {
			// 密码匹配
			return true, nil
		}
	} else if errors.Is(gorm.ErrRecordNotFound, queryErr) {
		// 用户不存在
		d.logger.LogModuleInfo(ctx, "user not exist", username)
		return false, nil
	} else {
		// 查询用户失败
		d.logger.LogModuleError(ctx, "failed to query user", queryErr)
		return false, errors.Derive("failed to query user", queryErr)
	}
}

// ActivateUser 激活用户
//   - username: 用户名
func (d *dao) ActivateUser(ctx context.Context, username string) (success bool, err errors.AliothError) {
	// 执行激活用户事务
	condition, updated := model.UserDTO{Username: username, Enable: false}, model.UserDTO{Enable: true}
	if activateErr := d.users.UpdateOneByCondition(condition, updated); activateErr == nil {
		// 激活用户成功
		d.logger.LogModuleInfo(ctx, "activate user success", username)
		return true, nil
	} else {
		// 激活用户失败
		d.logger.LogModuleError(ctx, "failed to activate user", activateErr)
		return false, activateErr
	}
}

// ActivatePhone 激活手机
//   - username: 用户名
//   - code: 验证码
//   - phone: 手机号
func (d *dao) ActivatePhone(ctx context.Context, username, phone string) (success bool, err errors.AliothError) {
	// 执行激活手机事务
	condition, updated := model.UserDTO{Username: username, Enable: true}, model.UserDTO{Phone: phone}
	if activateErr := d.users.UpdateOneByCondition(condition, updated); activateErr != nil {
		// 激活手机失败
		d.logger.LogModuleError(ctx, "failed to activate phone", activateErr)
		return false, activateErr
	} else {
		// 激活手机成功
		d.logger.LogModuleInfo(ctx, "activate phone success", username)
		return true, nil
	}
}

// UpdateUser 更新用户信息
//   - username: 用户名
//   - nickname: 昵称
//   - avatar: 头像url
func (d *dao) UpdateUser(ctx context.Context, username, nickname, avatar string) (success bool, err errors.AliothError) {
	// 更新用户信息
	condition, updated := model.UserDTO{Username: username, Enable: true}, model.UserDTO{Nickname: nickname, Avatar: avatar}
	if updateErr := d.users.UpdateOneByCondition(condition, updated); updateErr != nil {
		// 更新用户信息失败
		d.logger.LogModuleError(ctx, "failed to update user", updateErr)
		return false, updateErr
	} else {
		// 更新用户信息成功
		d.logger.LogModuleInfo(ctx, "update user success", username)
		return true, nil
	}
}

// UpdateUserPassword 更新用户密码
//   - username: 用户名
//   - newPass: 新密码
func (d *dao) UpdateUserPassword(ctx context.Context, username string, newPass string) (success bool, err errors.AliothError) {
	// 编码新密码
	if bcryptEncode, encodePasswdErr := passwd.BcryptEncode(newPass); encodePasswdErr != nil {
		// 密码编码失败
		d.logger.LogModuleError(ctx, "failed to encode password", encodePasswdErr)
		return false, errors.Derive("failed to encode password", encodePasswdErr)
	} else if updateErr := d.users.UpdateOneByCondition(model.UserDTO{Username: username, Enable: true}, model.UserDTO{Password: bcryptEncode}); updateErr != nil {
		// 更新用户密码失败
		d.logger.LogModuleError(ctx, "failed to update user password", updateErr)
		return false, updateErr
	} else {
		// 更新用户密码成功
		d.logger.LogModuleInfo(ctx, "update user password success", username)
		return true, nil
	}
}

// UpdateUserEmail 更新用户邮箱
//   - username: 用户名
//   - newEmail: 新邮箱
func (d *dao) UpdateUserEmail(ctx context.Context, username string, newEmail string) (success bool, err errors.AliothError) {
	// 不需要邮箱验证，直接更新邮箱
	if updateErr := d.users.UpdateOneByCondition(model.UserDTO{Username: username, Enable: true}, model.UserDTO{Email: newEmail}); updateErr != nil {
		// 更新用户邮箱失败
		d.logger.LogModuleError(ctx, "failed to update user email", updateErr)
		return false, updateErr
	} else {
		// 更新用户邮箱成功
		d.logger.LogModuleInfo(ctx, "update user email success", username)
		return true, nil
	}
}

// DisableUser 禁用用户
//   - username: 用户名
func (d *dao) DisableUser(ctx context.Context, username string) (success bool, err errors.AliothError) {
	// 禁用用户
	if updateErr := d.users.UpdateOneByCondition(model.UserDTO{Username: username, Enable: true}, model.UserDTO{Enable: false}); updateErr != nil {
		// 禁用用户失败
		d.logger.LogModuleError(ctx, "failed to disable user", updateErr)
		return false, updateErr
	} else {
		// 禁用用户成功
		d.logger.LogModuleInfo(ctx, "disable user success", username)
		return true, nil
	}
}

// DisableUserPhone 禁用用户手机
//   - username: 用户名
func (d *dao) DisableUserPhone(ctx context.Context, username string) (success bool, err errors.AliothError) {
	// 禁用用户手机
	if updateErr := d.users.UpdateOneByCondition(model.UserDTO{Username: username, Enable: true}, model.UserDTO{Phone: ""}); updateErr != nil {
		// 禁用用户手机失败
		d.logger.LogModuleError(ctx, "failed to disable user phone", updateErr)
		return false, updateErr
	} else {
		// 禁用用户手机成功
		d.logger.LogModuleInfo(ctx, "disable user phone success", username)
		return true, nil
	}
}

// GetUserInfo 获取用户信息
//   - username: 用户名
func (d *dao) GetUserInfo(ctx context.Context, username string) (user model.UserDTO, err errors.AliothError) {
	if queryResult, queryErr := d.users.CustomQueryOne("username = ?", username); queryErr != nil {
		// 查询用户失败
		d.logger.LogModuleError(ctx, "failed to query user", queryErr)
		return model.UserDTO{}, queryErr
	} else {
		// 查询用户成功
		d.logger.LogModuleInfo(ctx, "query user success", username)
		return queryResult, nil
	}
}

// GetUserInfoByEmail 获取用户信息
//   - email: 邮箱
func (d *dao) GetUserInfoByEmail(ctx context.Context, email string) (user model.UserDTO, err errors.AliothError) {
	if queryResult, queryErr := d.users.CustomQueryOne("email = ?", email); queryErr != nil {
		// 查询用户失败
		d.logger.LogModuleError(ctx, "failed to query user", queryErr)
		return model.UserDTO{}, queryErr
	} else {
		// 查询用户成功
		d.logger.LogModuleInfo(ctx, "query user success", email)
		return queryResult, nil
	}
}

// RemoveUser 删除用户
//   - username: 用户名
func (d *dao) RemoveUser(ctx context.Context, username string) (success bool, err errors.AliothError) {
	// 删除用户
	if deleteErr := d.users.DeleteOneByCondition(model.UserDTO{Username: username}); deleteErr != nil {
		// 删除用户失败
		d.logger.LogModuleError(ctx, "failed to delete user", deleteErr)
		return false, deleteErr
	} else {
		// 删除用户成功
		d.logger.LogModuleInfo(ctx, "delete user success", username)
		return true, nil
	}
}

// CreateApplication 创建应用
//   - appName: 应用名
//   - owner: 所有者
func (d *dao) CreateApplication(ctx context.Context, appName, owner, appKey, appSecret string) (application model.ApplicationDTO, err errors.AliothError) {
	// 装填应用信息
	application = model.ApplicationDTO{
		Name:      appName,
		Owner:     owner,
		Enable:    false,
		AppKey:    appKey,
		AppSecret: appSecret,
	}

	// 创建应用
	if createErr := d.applications.InsertOne(application); createErr != nil {
		// 创建应用失败
		d.logger.LogModuleError(ctx, "failed to create application", createErr)
		return model.ApplicationDTO{}, createErr
	} else {
		// 创建应用成功
		d.logger.LogModuleInfo(ctx, "create application success", appName)
		return application, nil
	}
}

// CheckCallbackMessage 检查回调消息是否匹配
//   - appName: 应用名
//   - code: 回调消息
func (d *dao) CheckCallbackMessage(ctx context.Context, appName, message string) (success bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisCallbackMessageKeyPrefix, appName)

	// 检查回调消息是否匹配
	if hasCorrectedCode, correctCode, getCorrectCodeErr := d.rds.Load(ctx, redisKey); getCorrectCodeErr != nil {
		// 获取回调消息失败
		d.logger.LogModuleError(ctx, "failed to get callback message", getCorrectCodeErr)
		return false, getCorrectCodeErr
	} else if !hasCorrectedCode {
		// 回调消息不存在
		d.logger.LogModuleInfo(ctx, "callback message not exist", appName)
		return false, nil
	} else if correctCode != message {
		// 回调消息不匹配
		d.logger.LogModuleInfo(ctx, "callback message not match", appName)
		return false, nil
	} else {
		// 验证回调消息成功
		d.logger.LogModuleInfo(ctx, "callback message match", appName)
		return true, nil
	}
}

// CheckApplicationSecret 检查应用密钥是否匹配
//   - appKey: 应用Key
//   - appSecret: 应用密钥
func (d *dao) CheckApplicationSecret(ctx context.Context, appKey, appSecret string) (success bool, err errors.AliothError) {
	if countOfApplication, countApplication := d.applications.CustomQueryCount("app_key = ? and app_secret = ?", appKey, appSecret); countApplication != nil {
		// 查询应用失败
		d.logger.LogModuleError(ctx, "failed to query application", countApplication)
		return false, countApplication
	} else {
		// 查询应用成功
		d.logger.LogModuleInfo(ctx, "query application success", appKey)
		return countOfApplication == 1, nil
	}
}

// CheckApplicationStatus 检查应用是否存在以及是否激活
//   - appName: 应用名
func (d *dao) CheckApplicationStatus(ctx context.Context, appName string) (exist bool, activated bool, err errors.AliothError) {
	// 查询应用是否激活
	if application, queryErr := d.applications.CustomQueryOne("name = ?", appName); queryErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, queryErr) {
			// 应用不存在
			d.logger.LogModuleInfo(ctx, "application not exist", appName)
			return false, false, nil
		} else {
			// 查询应用失败
			d.logger.LogModuleError(ctx, "failed to query application", queryErr)
			return false, false, queryErr
		}
	} else {
		// 查询应用成功
		d.logger.LogModuleInfo(ctx, "query application success", appName)
		return true, application.Enable, nil
	}
}

// ActivateApplicationCallback 激活应用回调地址
//   - appName: 应用名
//   - callback: 回调地址
func (d *dao) ActivateApplicationCallback(ctx context.Context, appName, callback string) (success bool, err errors.AliothError) {
	// 更新应用回调地址
	if updateErr := d.applications.UpdateOneByCondition(model.ApplicationDTO{Name: appName, Enable: false}, model.ApplicationDTO{Callback: callback, Enable: true}); updateErr != nil {
		// 更新应用回调地址失败
		d.logger.LogModuleError(ctx, "failed to update application callback", updateErr)
		return false, updateErr
	} else {
		// 更新应用回调地址成功
		d.logger.LogModuleInfo(ctx, "update application callback success", appName)
		return true, nil
	}
}

// UpdateApplicationCallback 更新应用回调地址
//   - appName: 应用名
//   - callback: 回调地址
func (d *dao) UpdateApplicationCallback(ctx context.Context, appName, callback string) (success bool, err errors.AliothError) {
	// 更新应用回调地址
	if updateErr := d.applications.UpdateOneByCondition(model.ApplicationDTO{Name: appName}, model.ApplicationDTO{Callback: callback, Enable: false}); updateErr != nil {
		// 更新应用回调地址失败
		d.logger.LogModuleError(ctx, "failed to update application callback", updateErr)
		return false, updateErr
	} else {
		// 更新应用回调地址成功
		d.logger.LogModuleInfo(ctx, "update application callback success", appName)
		return true, nil
	}
}

// UpdateApplicationInfo 更新应用信息
//   - appName: 应用名
//   - avatar: 头像url
//   - website: 网站url
//   - email: 邮箱
func (d *dao) UpdateApplicationInfo(ctx context.Context, appName, avatar, website, email string) (success bool, err errors.AliothError) {
	// 更新应用信息
	if updateErr := d.applications.UpdateOneByCondition(model.ApplicationDTO{Name: appName}, model.ApplicationDTO{Avatar: avatar, Website: website, Email: email}); updateErr != nil {
		// 更新应用信息失败
		d.logger.LogModuleError(ctx, "failed to update application info", updateErr)
		return false, updateErr
	} else {
		// 更新应用信息成功
		d.logger.LogModuleInfo(ctx, "update application info success", appName)
		return true, nil
	}
}

// UpdateApplicationSecret 更新应用密钥
//   - appName: 应用名
//   - appKey: 应用Key
//   - appSecret: 应用密钥
func (d *dao) UpdateApplicationSecret(ctx context.Context, appName, appKey, appSecret string) (success bool, err errors.AliothError) {
	// 更新应用密钥
	if updateErr := d.applications.UpdateOneByCondition(model.ApplicationDTO{Name: appName, AppKey: appKey}, model.ApplicationDTO{AppSecret: appSecret}); updateErr != nil {
		// 更新应用密钥失败
		d.logger.LogModuleError(ctx, "failed to update application secret", updateErr)
		return false, updateErr
	} else {
		// 更新应用密钥成功
		d.logger.LogModuleInfo(ctx, "update application secret success", appName)
		return true, nil
	}
}

// GetApplicationInfo 获取应用信息
//   - appName: 应用名
func (d *dao) GetApplicationInfo(ctx context.Context, appName string) (application model.ApplicationDTO, err errors.AliothError) {
	if applicationResult, getApplicationErr := d.applications.CustomQueryOne("name = ?", appName); getApplicationErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, getApplicationErr) {
			// 应用不存在
			d.logger.LogModuleInfo(ctx, "application not exist", appName)
			return model.ApplicationDTO{}, errors.NewApplicationNotExistsError(appName)
		} else {
			// 查询应用失败
			d.logger.LogModuleError(ctx, "failed to query application", getApplicationErr)
			return model.ApplicationDTO{}, getApplicationErr
		}
	} else {
		// 查询应用成功
		d.logger.LogModuleInfo(ctx, "query application success", appName)
		return applicationResult, nil
	}
}

// GetApplicationInfoByAppKey 获取应用信息
//   - appKey: 应用Key
func (d *dao) GetApplicationInfoByAppKey(ctx context.Context, appKey string) (application model.ApplicationDTO, err errors.AliothError) {
	if applicationResult, getApplicationErr := d.applications.CustomQueryList("app_key = ?", appKey); getApplicationErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, getApplicationErr) {
			// 应用不存在
			d.logger.LogModuleInfo(ctx, "application not exist", appKey)
			return model.ApplicationDTO{}, errors.NewApplicationNotExistsError(appKey)
		} else {
			// 查询应用失败
			d.logger.LogModuleError(ctx, "failed to query application", getApplicationErr)
			return model.ApplicationDTO{}, getApplicationErr
		}
	} else {
		// 查询应用成功
		d.logger.LogModuleInfo(ctx, "query application success", appKey)
		return applicationResult[0], nil
	}
}

// RemoveApplication 删除应用
//   - appName: 应用名
func (d *dao) RemoveApplication(ctx context.Context, appName string) (success bool, err errors.AliothError) {
	// 删除应用
	if deleteErr := d.applications.DeleteOneByCondition(model.ApplicationDTO{Name: appName}); deleteErr != nil {
		// 删除应用失败
		d.logger.LogModuleError(ctx, "failed to delete application", deleteErr)
		return false, deleteErr
	} else {
		// 删除应用成功
		d.logger.LogModuleInfo(ctx, "delete application success", appName)
		return true, nil
	}
}

// CreateScope 创建权限
//   - scopeName: 权限名
//   - description: 权限描述
func (d *dao) CreateScope(ctx context.Context, scopeName, description, application string) (scope model.ScopeDTO, err errors.AliothError) {
	// 装填权限信息
	scope = model.ScopeDTO{
		Name:        scopeName,
		Application: application,
		Description: description,
	}

	// 创建权限
	if createErr := d.scopes.InsertOne(scope); createErr != nil {
		// 创建权限失败
		d.logger.LogModuleError(ctx, "failed to create scope", createErr)
		return model.ScopeDTO{}, createErr
	} else {
		// 创建权限成功
		d.logger.LogModuleInfo(ctx, "create scope success", scopeName)
		return scope, nil
	}
}

// UpdateScope 更新权限
//   - scopeName: 权限名
//   - scopeDescription: 权限描述
func (d *dao) UpdateScope(ctx context.Context, scopeName, scopeDescription string) (success bool, err errors.AliothError) {
	// 更新权限
	if updateErr := d.scopes.UpdateOneByCondition(model.ScopeDTO{Name: scopeName}, model.ScopeDTO{Description: scopeDescription}); updateErr != nil {
		// 更新权限失败
		d.logger.LogModuleError(ctx, "failed to update scope", updateErr)
		return false, updateErr
	} else {
		// 更新权限成功
		d.logger.LogModuleInfo(ctx, "update scope success", scopeName)
		return true, nil
	}
}

// GetScopeInfo 获取权限信息
//   - scopeName: 权限名
func (d *dao) GetScopeInfo(ctx context.Context, scopeName string) (scope model.ScopeDTO, err errors.AliothError) {
	if scopeResult, getScopeErr := d.scopes.CustomQueryOne("name = ?", scopeName); getScopeErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, getScopeErr) {
			// 权限不存在
			d.logger.LogModuleInfo(ctx, "scope not exist", scopeName)
			return model.ScopeDTO{}, errors.NewScopeNotExistsError(scopeName)
		} else {
			// 查询权限失败
			d.logger.LogModuleError(ctx, "failed to query scope", getScopeErr)
			return model.ScopeDTO{}, getScopeErr
		}
	} else {
		// 查询权限成功
		d.logger.LogModuleInfo(ctx, "query scope success", scopeName)
		return scopeResult, nil
	}
}

// GetScopeList 获取权限列表
//   - application: 应用名
func (d *dao) GetScopeList(ctx context.Context, application string) (scopes []model.ScopeDTO, err errors.AliothError) {
	if scopeResult, getScopeErr := d.scopes.CustomQueryList("application = ?", application); getScopeErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, getScopeErr) {
			// 权限不存在
			d.logger.LogModuleInfo(ctx, "scope not exist", application)
			return []model.ScopeDTO{}, nil
		} else {
			// 查询权限失败
			d.logger.LogModuleError(ctx, "failed to query scope", getScopeErr)
			return []model.ScopeDTO{}, getScopeErr
		}
	} else {
		// 查询权限成功
		d.logger.LogModuleInfo(ctx, "query scope success", application)
		return scopeResult, nil
	}
}

// GetScopeListByNames 获取权限列表
//   - scopeNames: 权限名列表
func (d *dao) GetScopeListByNames(ctx context.Context, scopeNames ...string) (scopes []model.ScopeDTO, err errors.AliothError) {
	// 检查权限名列表是否为空
	if len(scopeNames) == 0 {
		return []model.ScopeDTO{}, nil
	}

	// 查询权限
	if scopeResult, getScopeErr := d.scopes.CustomQueryList("name in (?)", scopeNames); getScopeErr != nil {
		if errors.Is(gorm.ErrRecordNotFound, getScopeErr) {
			// 权限不存在
			d.logger.LogModuleInfo(ctx, "scope not exist", scopeNames)
			return []model.ScopeDTO{}, nil
		} else {
			// 查询权限失败
			d.logger.LogModuleError(ctx, "failed to query scope", getScopeErr)
			return []model.ScopeDTO{}, getScopeErr
		}
	} else {
		// 查询权限成功
		d.logger.LogModuleInfo(ctx, "query scope success", scopeNames)
		return scopeResult, nil
	}
}

// RemoveScope 删除权限
//   - scopeName: 权限名
func (d *dao) RemoveScope(ctx context.Context, scopeName string) (success bool, err errors.AliothError) {
	// 删除权限
	if deleteErr := d.scopes.DeleteOneByCondition(model.ScopeDTO{Name: scopeName}); deleteErr != nil {
		// 删除权限失败
		d.logger.LogModuleError(ctx, "failed to delete scope", deleteErr)
		return false, deleteErr
	} else {
		// 删除权限成功
		d.logger.LogModuleInfo(ctx, "delete scope success", scopeName)
		return true, nil
	}
}

// RemoveScopeByApplication 删除应用下的所有权限
//   - application: 应用名
func (d *dao) RemoveScopeByApplication(ctx context.Context, application string) (success bool, err errors.AliothError) {
	// 删除权限
	if deleteErr := d.scopes.DeleteListByCondition(model.ScopeDTO{Application: application}); deleteErr != nil {
		// 删除权限失败
		d.logger.LogModuleError(ctx, "failed to delete scope", deleteErr)
		return false, deleteErr
	} else {
		// 删除权限成功
		d.logger.LogModuleInfo(ctx, "delete scope success", application)
		return true, nil
	}
}

// StoreCode2TokenToken 将OAuth的code2token环节的令牌存储
//   - code: 令牌键
//   - token: 令牌
func (d *dao) StoreCode2TokenToken(ctx context.Context, code, token string) (success bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisCode2TokenKeyPrefix, code)
	redisExpiredTime := time.Duration(initialize.GlobalConfig().Starward.Security.TokenTimeoutSeconds) * time.Second

	// 存储令牌
	if storeErr := d.rds.StoreEX(ctx, redisKey, token, redisExpiredTime); storeErr != nil {
		// 存储令牌失败
		d.logger.LogModuleError(ctx, "failed to store token", storeErr)
		return false, storeErr
	} else {
		// 存储令牌成功
		d.logger.LogModuleInfo(ctx, "store token success", code)
		return true, nil
	}
}

// GetCode2TokenToken 获取OAuth的code2token环节的令牌
//   - code: 令牌键
func (d *dao) GetCode2TokenToken(ctx context.Context, code string) (token string, expiredTime time.Duration, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisCode2TokenKeyPrefix, code)

	// 获取令牌
	if loaded, expired, token, loadErr := d.rds.LoadWithEX(ctx, redisKey); loadErr != nil {
		// 获取令牌失败
		d.logger.LogModuleError(ctx, "failed to get token", loadErr)
		return "", 0, loadErr
	} else if !loaded {
		// 令牌不存在
		d.logger.LogModuleInfo(ctx, "token not exist", code)
		return "", 0, nil
	} else {
		// 获取令牌成功
		d.logger.LogModuleInfo(ctx, "get token success", code)
		return token, expired, nil
	}
}

// RemoveCode2TokenToken 删除OAuth的code2token环节的令牌
//   - code: 令牌键
func (d *dao) RemoveCode2TokenToken(ctx context.Context, code string) (success bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisCode2TokenKeyPrefix, code)

	// 删除令牌
	if deleteErr := d.rds.Delete(ctx, redisKey); deleteErr != nil {
		// 删除令牌失败
		d.logger.LogModuleError(ctx, "failed to delete token", deleteErr)
		return false, deleteErr
	} else {
		// 删除令牌成功
		d.logger.LogModuleInfo(ctx, "delete token success", code)
		return true, nil
	}
}

// StoreScopeToken 将OAuth的token环节的令牌存储
//   - token: 令牌键
//   - scopes: 权限列表
func (d *dao) StoreScopeToken(ctx context.Context, token string, scopes ...string) (success bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisTokenScopesKeyPrefix, token)
	redisExpiredTime := time.Duration(initialize.GlobalConfig().Starward.Security.TokenTimeoutSeconds) * time.Second

	// 存储令牌
	if storeErr := d.rds.AddMembers(ctx, redisKey, scopes...); storeErr != nil {
		// 存储令牌失败
		d.logger.LogModuleError(ctx, "failed to store token", storeErr)
		return false, storeErr
	} else if setExpireErr := d.rds.Expire(ctx, redisKey, redisExpiredTime); setExpireErr != nil {
		// 设置令牌过期时间失败
		d.logger.LogModuleError(ctx, "failed to set token expire time", setExpireErr)
		return false, setExpireErr
	} else {
		// 存储令牌成功
		d.logger.LogModuleInfo(ctx, "store token success", token)
		return true, nil
	}
}

// CheckScopeTokenHasScopes 获取OAuth的token环节的令牌
//   - token: 令牌键
//   - scopes: 权限列表
func (d *dao) CheckScopeTokenHasScopes(ctx context.Context, token string, scopes ...string) (correct bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKey(redisTokenScopesKeyPrefix, token)

	if isMember, getIsMemberErr := d.rds.IsMembers(ctx, redisKey, scopes...); getIsMemberErr != nil {
		// 获取令牌权限失败
		d.logger.LogModuleError(ctx, "failed to get token scopes", getIsMemberErr)
		return false, getIsMemberErr
	} else {
		// 获取令牌权限成功
		d.logger.LogModuleInfo(ctx, "get token scopes success", token)
		return isMember, nil
	}
}

// RemoveScopeToken 删除OAuth的token环节的令牌
//   - token: 令牌键
func (d *dao) RemoveScopeToken(ctx context.Context, token string) (err errors.AliothError) {
	redisKey := utils.BuildRedisKey(redisTokenScopesKeyPrefix, token)

	if deleteErr := d.rds.Delete(ctx, redisKey); deleteErr != nil {
		// 删除令牌失败
		d.logger.LogModuleError(ctx, "failed to delete token", deleteErr)
		return deleteErr
	} else {
		// 删除令牌成功
		d.logger.LogModuleInfo(ctx, "delete token success", token)
		return nil
	}
}

// StoreUserInfoToken 将OAuth的token环节的令牌与用户信息绑定
//   - token: 令牌键
//   - username: 用户名
func (d *dao) StoreUserInfoToken(ctx context.Context, token, username string) (success bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisUserInfoScopeKeyPrefix, token)
	redisExpiredTime := time.Duration(initialize.GlobalConfig().Starward.Security.TokenTimeoutSeconds) * time.Second

	// 存储令牌与用户信息绑定
	if storeErr := d.rds.StoreEX(ctx, redisKey, username, redisExpiredTime); storeErr != nil {
		// 存储令牌与用户信息绑定失败
		d.logger.LogModuleError(ctx, "failed to store token user info", storeErr)
		return false, storeErr
	} else {
		// 存储令牌与用户信息绑定成功
		d.logger.LogModuleInfo(ctx, "store token user info success", token)
		return true, nil
	}
}

// GetUserInfoToken 获取OAuth的token环节的令牌与用户信息绑定
//   - token: 令牌键
func (d *dao) GetUserInfoToken(ctx context.Context, token string) (username string, expiredTime time.Duration, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisUserInfoScopeKeyPrefix, token)

	// 获取令牌与用户信息绑定
	if loaded, expired, username, loadErr := d.rds.LoadWithEX(ctx, redisKey); loadErr != nil {
		// 获取令牌与用户信息绑定失败
		d.logger.LogModuleError(ctx, "failed to get token user info", loadErr)
		return "", 0, loadErr
	} else if !loaded {
		// 令牌与用户信息绑定不存在
		d.logger.LogModuleInfo(ctx, "token user info not exist", token)
		return "", 0, nil
	} else {
		// 获取令牌与用户信息绑定成功
		d.logger.LogModuleInfo(ctx, "get token user info success", token)
		return username, expired, nil
	}
}

// RemoveUserInfoToken 删除OAuth的token环节的令牌与用户信息绑定
//   - token: 令牌键
func (d *dao) RemoveUserInfoToken(ctx context.Context, token string) (success bool, err errors.AliothError) {
	redisKey := utils.BuildRedisKeyWithoutPrefix(redisUserInfoScopeKeyPrefix, token)

	// 删除令牌与用户信息绑定
	if deleteErr := d.rds.Delete(ctx, redisKey); deleteErr != nil {
		// 删除令牌与用户信息绑定失败
		d.logger.LogModuleError(ctx, "failed to delete token user info", deleteErr)
		return false, deleteErr
	} else {
		// 删除令牌与用户信息绑定成功
		d.logger.LogModuleInfo(ctx, "delete token user info success", token)
		return true, nil
	}
}
