package model

import (
	"api/internal/user/domain"
	"shared/common"
)

func UserModelToDomain(user UserModel) (domain.User, error) {
	m, err := common.NewEmail(user.Email)
	if err != nil {
		return domain.User{}, err
	}
	p := domain.PasswordHashed(user.Password)
	r, err := domain.NewRole(user.Role)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:           user.Id,
		Email:        m,
		PasswordHash: p,
		Role:         r,
	}, nil
}
