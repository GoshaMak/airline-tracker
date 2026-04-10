package usecase

import (
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"fmt"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(i do.Injector) (*UserUsecase, error) {
	return &UserUsecase{
		repo: do.MustInvokeAs[repository.UserRepository](i),
	}, nil
}

func (uc *UserUsecase) Get(email, password string) (*domain.User, error) {
	op := "user_service.GetUser"
	var u domain.User
	var err error
	if email != "" {
		u, err = uc.repo.GetUser(context.Background(), email)
	} else {
		err = fmt.Errorf("Empty email and phone")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return nil, fmt.Errorf("%s: %s", op, "wrong password")
	}
	return &u, nil
}

func (uc *UserUsecase) Exist(email, password string) bool {
	u, err := uc.Get(email, password)
	if err != nil || u == nil {
		return false
	}
	return true
}
