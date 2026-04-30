package model

import (
	"airline-tracker/internal/common"
	"airline-tracker/internal/user/domain"
)

func UserModelToDomain(um UserModel) (domain.User, error) {
	m, err := common.NewEmail(um.Email)
	if err != nil {
		return domain.User{}, err
	}
	p := domain.PasswordHashed(um.Password)
	r, err := common.NewRole(um.Role)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		ID:           um.ID,
		Email:        m,
		PasswordHash: p,
		Role:         r,
	}, nil
}
