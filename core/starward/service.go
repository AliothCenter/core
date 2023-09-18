package starward

import (
	"context"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

type Service interface {
}

type service struct {
	dao *dao
}

func (s *service) UserRegistration(ctx context.Context, request *alioth.UserRegistrationRequest) (*alioth.UserRegistrationResponse, error) {
	// 检查用户是否存在
	if existUser, _, checkExistUserErr := s.dao.CheckUserStatus(ctx, request.GetUsername()); checkExistUserErr != nil {
		return nil, errors.Derive("check user exist error", checkExistUserErr)
	} else if existUser {
		return nil, errors.NewUserAlreadyExistsError(request.GetUsername())
	}

	// 检查邮箱是否已经使用
	if inuse, checkEmailInUseErr := s.dao.CheckEmailInuse(ctx, request.GetEmail()); checkEmailInUseErr != nil {
		return nil, errors.Derive("check email in use error", checkEmailInUseErr)
	} else if inuse {
		return nil, errors.NewEmailAlreadyInUseError(request.GetEmail())
	}

	// 发送注册邮件
	// todo: implement me!
	panic("implement me!")
}
