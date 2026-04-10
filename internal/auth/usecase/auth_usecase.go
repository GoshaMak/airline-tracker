package usecase

import (
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"fmt"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo repository.UserRepository
}

func NewAuthUsecase(i do.Injector) (*AuthUsecase, error) {
	return &AuthUsecase{
		repo: do.MustInvokeAs[repository.UserRepository](i),
	}, nil
}

func (uc *AuthUsecase) GetUser(email, password string) (*domain.User, error) {
	op := "auth_usecase.GetUser"

	u, err := uc.repo.GetUser(context.Background(), email)
	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return nil, fmt.Errorf("%s: %s", op, "wrong password")
	}

	return &u, nil
}

func (uc *AuthUsecase) CreateUser(u *domain.User) error {
	if encryptPassword(u) != nil { // TODO: move inside NewUser
		return fmt.Errorf("Can't encrypt password")
	}

	if err := uc.repo.Save(context.Background(), u); err != nil {
		return err
	}

	return nil
}

func encryptPassword(u *domain.User) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (uc *AuthUsecase) Exists(email, password string) bool {
	u, err := uc.GetUser(email, password)
	if err != nil || u == nil {
		return false
	}
	return true
}
