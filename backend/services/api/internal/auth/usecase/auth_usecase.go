package usecase

import (
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"context"
	"errors"
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

func (uc *AuthUsecase) GetUser(email, password string) (domain.User, error) {
	op := "AuthUsecase.GetUser"
	u, err := uc.repo.GetUser(context.Background(), email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (uc *AuthUsecase) CreateUser(u domain.User) error {
	op := "AuthUsecase.CreateUser"

	if err := uc.repo.SaveUser(context.Background(), u); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (uc *AuthUsecase) Exists(email, password string) bool {
	_, err := uc.GetUser(email, password)
	if err != nil { // WARN:
		return false
	}
	return true
}
